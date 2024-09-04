package httpsling

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"syscall"
	"time"
)

// DefaultRetryConfig is the default retry configuration used if nil is passed to Retry()
var DefaultRetryConfig = RetryConfig{}

// DefaultBackoff is a backoff configuration with the default values
var DefaultBackoff = ExponentialBackoff{
	BaseDelay:  1.0 * time.Second, // nolint: mnd
	Multiplier: 1.6,               // nolint: mnd
	Jitter:     0.2,               // nolint: mnd
	MaxDelay:   120 * time.Second, // nolint: mnd
}

// DefaultShouldRetry is the default ShouldRetryer
func DefaultShouldRetry(_ int, _ *http.Request, resp *http.Response, err error) bool {
	var netError net.Error

	switch {
	case err == nil:
		return resp.StatusCode == 500 || resp.StatusCode > 501 || resp.StatusCode == 429
	case errors.Is(err, io.EOF),
		errors.Is(err, syscall.ECONNRESET),
		errors.Is(err, syscall.ECONNABORTED),
		errors.Is(err, syscall.EPIPE):
		return true
	case errors.As(err, &netError) && netError.Timeout():
		return true
	}

	return false
}

// OnlyIdempotentShouldRetry returns true if the request is using one of the HTTP methods which are intended to be idempotent: GET, HEAD, OPTIONS, and TRACE
func OnlyIdempotentShouldRetry(_ int, req *http.Request, _ *http.Response, _ error) bool {
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// RetryConfig defines settings for the Retry middleware
type RetryConfig struct {
	// MaxAttempts is the number of times to attempt the request (default 3)
	MaxAttempts int
	// ShouldRetry tests whether a response should be retried
	ShouldRetry ShouldRetryer
	// Backoff returns how long to wait between retries
	Backoff Backoffer
	// ReadResponse will ensure the entire response is read before considering the request a success
	ReadResponse bool
}

func (c *RetryConfig) normalize() {
	if c.Backoff == nil {
		c.Backoff = &DefaultBackoff
	}

	if c.ShouldRetry == nil {
		c.ShouldRetry = ShouldRetryerFunc(DefaultShouldRetry)
	}

	if c.MaxAttempts < 1 {
		c.MaxAttempts = 3
	}
}

// ShouldRetryer evaluates whether an HTTP request should be retried
type ShouldRetryer interface {
	ShouldRetry(attempt int, req *http.Request, resp *http.Response, err error) bool
}

// ShouldRetryerFunc adapts a function to the ShouldRetryer interface
type ShouldRetryerFunc func(attempt int, req *http.Request, resp *http.Response, err error) bool

// ShouldRetry implements ShouldRetryer
func (s ShouldRetryerFunc) ShouldRetry(attempt int, req *http.Request, resp *http.Response, err error) bool {
	return s(attempt, req, resp, err)
}

// AllRetryers returns a ShouldRetryer which returns true only if all the supplied retryers return true
func AllRetryers(s ...ShouldRetryer) ShouldRetryer {
	return ShouldRetryerFunc(func(attempt int, req *http.Request, resp *http.Response, err error) bool {
		for _, shouldRetryer := range s {
			if !shouldRetryer.ShouldRetry(attempt, req, resp, err) {
				return false
			}
		}

		return true
	})
}

// Backoffer calculates how long to wait between attempts
type Backoffer interface {
	Backoff(attempt int) time.Duration
}

// BackofferFunc adapts a function to the Backoffer interface
type BackofferFunc func(int) time.Duration

// Backoff implements Backoffer
func (b BackofferFunc) Backoff(attempt int) time.Duration {
	return b(attempt)
}

// ExponentialBackoff defines the configuration options for an exponential backoff strategy
type ExponentialBackoff struct {
	// BaseDelay is the amount of time to backoff after the first failure
	BaseDelay time.Duration
	// Multiplier is the factor with which to multiply backoffs after a failed retry
	Multiplier float64
	// Jitter is the factor with which backoffs are randomized
	Jitter float64
	// MaxDelay is the upper bound of backoff delay - 0 means no max
	MaxDelay time.Duration
}

func (c *ExponentialBackoff) Backoff(attempt int) time.Duration {
	backoff := float64(c.BaseDelay)

	if c.Multiplier > 0 {
		backoff *= math.Pow(c.Multiplier, float64(attempt-1))
	}

	maxDelayf := float64(c.MaxDelay)
	if c.MaxDelay > 0 {
		backoff = math.Min(backoff, maxDelayf)
	}

	backoff = math.Max(0, backoff)

	if c.Jitter > 0 {
		// nolint:gosec
		backoff *= 1 + c.Jitter*(rand.Float64()*2-1)
		if c.MaxDelay > 0 {
			if delta := backoff - maxDelayf; delta > 0 {
				// jitter bumped the backoff above max delay.  Redistribute
				// below max
				backoff = maxDelayf - delta
			}
		}
	}

	return time.Duration(backoff)
}

// NoBackoff returns a Backoffer with zero backoff, and zero delay between retries
func NoBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{}
}

// ConstantBackoff returns a Backoffer with a fixed, constant delay between retries and no jitter
func ConstantBackoff(delay time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{BaseDelay: delay}
}

// ConstantBackoffWithJitter returns a Backoffer with a fixed, constant delay between retries with 20% jitter
func ConstantBackoffWithJitter(delay time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{BaseDelay: delay, Jitter: 0.2} // nolint: mnd
}

// Retry retries the http request under certain conditions - the number of retries,
// retry conditions, and the time to sleep between retries can be configured
func Retry(config *RetryConfig) Middleware {
	c := DefaultRetryConfig
	if config != nil {
		c = *config
	}

	c.normalize()

	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			if bodyEmpty(req) {
				return next.Do(req)
			}

			var (
				resp    *http.Response
				err     error
				attempt int
			)

			for {
				resp, err = next.Do(req)
				attempt++

				if err == nil && c.ReadResponse {
					resp.Body, err = bufRespBody(resp.Body)
				}

				if attempt >= c.MaxAttempts || !c.ShouldRetry.ShouldRetry(attempt, req, resp, err) {
					break
				}

				if resp != nil {
					drain(resp.Body)
				}

				req, err = resetRequest(req)
				if err != nil {
					return resp, err
				}

				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(c.Backoff.Backoff(attempt)):
				}
			}

			return resp, err
		})
	}
}

func bodyEmpty(req *http.Request) bool {
	return req.Body != nil && req.Body != http.NoBody && req.GetBody == nil
}

type errCloser struct {
	io.Reader
	err error
}

func (e *errCloser) Close() error {
	return e.err
}

// bufRespBody reads all of b to memory and then returns a ReadCloser with the same bytes
func bufRespBody(b io.ReadCloser) (r io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		return b, nil
	}

	var buf bytes.Buffer

	if _, err = buf.ReadFrom(b); err != nil {
		return nil, err
	}

	if err := b.Close(); err != nil {
		return &errCloser{
			Reader: &buf,
			err:    err,
		}, nil
	}

	return io.NopCloser(&buf), nil
}

func resetRequest(req *http.Request) (*http.Request, error) {
	copyReq := *req
	req = &copyReq

	if req.Body != nil && req.Body != http.NoBody {
		b, err := req.GetBody()
		if err != nil {
			return nil, fmt.Errorf("error calling req.GetBody: %w", err)
		}

		req.Body = b
	}

	return req, nil
}

func drain(r io.ReadCloser) {
	if r == nil {
		return
	}
	defer func(r io.ReadCloser) {
		_ = r.Close()
	}(r)

	_, _ = io.Copy(io.Discard, io.LimitReader(r, 4096)) // nolint: mnd
}
