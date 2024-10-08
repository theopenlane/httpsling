package httptestutil

import (
	"bytes"
	"io"
	"net/http"

	"github.com/felixge/httpsnoop"
)

// Exchange is a snapshot of one request/response exchange with the server
type Exchange struct {
	// Request is the request sent to the server
	Request *http.Request
	// RequestBody is the request body
	RequestBody *bytes.Buffer
	// StatusCode is the status code returned by the server
	StatusCode int
	// Header is the response header
	Header http.Header
	// ResponseBody is the response body
	ResponseBody *bytes.Buffer
}

// Inspector is server-side middleware which captures server exchanges in a buffer
type Inspector struct {
	Exchanges chan Exchange
}

// NewInspector creates a new Inspector with the requested channel buffer size
func NewInspector(size int) *Inspector {
	if size == 0 {
		size = 50
	}

	return &Inspector{
		Exchanges: make(chan Exchange, size),
	}
}

// NextExchange receives the next exchange from the channel, or returns nil if no exchange is ready
func (b *Inspector) NextExchange() *Exchange {
	select {
	case e := <-b.Exchanges:
		return &e
	default:
		return nil
	}
}

// LastExchange receives the most recent exchange from channel; this also has the side effect of draining the channel completely
func (b *Inspector) LastExchange() *Exchange {
	var e *Exchange

	for {
		select {
		case ex := <-b.Exchanges:
			e = &ex
		default:
			return e
		}
	}
}

// Drain reads all buffered exchanges from the channel
func (b *Inspector) Drain() []*Exchange {
	var e []*Exchange

	for {
		select {
		case ex := <-b.Exchanges:
			e = append(e, &ex)
		default:
			return e
		}
	}
}

// Clear drains the channel
func (b *Inspector) Clear() {
	if b == nil {
		return
	}

	b.LastExchange()
}

// Wrap installs the inspector in an HTTP server by wrapping the server's Handler
func (b *Inspector) Wrap(next http.Handler) http.Handler {
	if next == nil {
		next = http.DefaultServeMux
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ex := Exchange{}
		ex.Request = r

		if r.Body != nil && r.Body != http.NoBody {
			ex.RequestBody = &bytes.Buffer{}
			if _, err := ex.RequestBody.ReadFrom(r.Body); err != nil {
				panic(err)
			}

			if err := r.Body.Close(); err != nil {
				panic(err)
			}

			r.Body = io.NopCloser(bytes.NewReader(ex.RequestBody.Bytes()))
		} else {
			ex.RequestBody = nil
		}

		w = httpsnoop.Wrap(w, hooks(&ex))

		next.ServeHTTP(w, r)

		select {
		case b.Exchanges <- ex:
		default:
			// don't block if channel is full, just drop
		}
	})
}

func hooks(ex *Exchange) httpsnoop.Hooks {
	if ex.ResponseBody == nil {
		ex.ResponseBody = &bytes.Buffer{}
	}

	return httpsnoop.Hooks{
		Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(b []byte) (int, error) {
				ex.ResponseBody.Write(b)

				return next(b)
			}
		},
		Header: func(next httpsnoop.HeaderFunc) httpsnoop.HeaderFunc {
			return func() http.Header {
				ex.Header = next()

				return ex.Header
			}
		},
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(code int) {
				ex.StatusCode = code
				next(code)
			}
		},
		ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
			return func(src io.Reader) (int64, error) {
				l := ex.ResponseBody.Len()
				n, err := ex.ResponseBody.ReadFrom(src)
				if err != nil {
					return n, err
				}

				return next(bytes.NewReader(ex.ResponseBody.Bytes()[l:]))
			}
		},
	}
}
