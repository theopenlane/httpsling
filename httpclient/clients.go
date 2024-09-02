package httpclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/ansel1/merry"
)

// New builds a new *http.Client
func New(opts ...Option) (*http.Client, error) {
	c := &http.Client{}
	return c, Apply(c, opts...)
}

// Apply applies options to an existing client
func Apply(c *http.Client, opts ...Option) error {
	for _, opt := range opts {
		err := opt.Apply(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func newDefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // nolint: mnd
			KeepAlive: 30 * time.Second, // nolint: mnd
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,              // nolint: mnd
		IdleConnTimeout:       90 * time.Second, // nolint: mnd
		TLSHandshakeTimeout:   10 * time.Second, // nolint: mnd
		ExpectContinueTimeout: 1 * time.Second,  // nolint: mnd
	}
}

// Option is a configuration option for building an http.Client
type Option interface {
	Apply(*http.Client) error
}

// OptionFunc adapts a function to the Option interface
type OptionFunc func(*http.Client) error

// Apply implements Option
func (f OptionFunc) Apply(c *http.Client) error {
	return f(c)
}

// TransportOption configures the client's transport
type TransportOption func(transport *http.Transport) error

// Apply implements Option
func (f TransportOption) Apply(c *http.Client) error {
	var transport *http.Transport

	rt := c.Transport

	switch t := rt.(type) {
	case nil:
		transport = newDefaultTransport()
		c.Transport = transport
	case *http.Transport:
		transport = t
	default:
		return merry.Errorf("client.Transport is not a *http.Transport.  It's a %T", c.Transport)
	}

	return f(transport)
}

// A TLSOption is a type of Option which configures the TLS configuration of the client
type TLSOption func(c *tls.Config) error

// Apply implements Option
func (f TLSOption) Apply(c *http.Client) error {
	return TransportOption(func(t *http.Transport) error {
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		return f(t.TLSClientConfig)
	}).Apply(c)
}
