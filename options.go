package httpsling

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	goquery "github.com/google/go-querystring/query"

	"github.com/theopenlane/httpsling/httpclient"
)

// Option applies some setting to a Requester object
type Option interface {
	Apply(*Requester) error
}

// OptionFunc adapts a function to the Option interface
type OptionFunc func(*Requester) error

// Apply implements Option
func (f OptionFunc) Apply(r *Requester) error {
	return f(r)
}

// With clones the Requester, then applies the options to the clone
func (r *Requester) With(opts ...Option) (*Requester, error) {
	r2 := r.Clone()

	err := r2.Apply(opts...)
	if err != nil {
		return nil, err
	}

	return r2, nil
}

// MustWith clones the Requester, then applies the options to the clone
func (r *Requester) MustWith(opts ...Option) *Requester {
	requester, err := r.With(opts...)
	if err != nil {
		panic(err)
	}

	return requester
}

// Apply applies the options to the receiver
func (r *Requester) Apply(opts ...Option) error {
	for _, o := range opts {
		if err := o.Apply(r); err != nil {
			return fmt.Errorf("error applying options: %w", err)
		}
	}

	return nil
}

// MustApply applies the options to the receiver
func (r *Requester) MustApply(opts ...Option) {
	if err := r.Apply(opts...); err != nil {
		panic(err)
	}
}

// Method sets the HTTP method
func Method(m string, paths ...string) Option {
	return OptionFunc(func(r *Requester) error {
		r.Method = m

		if len(paths) == 0 {
			return nil
		}

		return RelativeURL(paths...).Apply(r)
	})
}

// Head sets the HTTP method to "HEAD"
func Head(paths ...string) Option {
	return Method(http.MethodHead, paths...)
}

// Get sets the HTTP method to "GET"
func Get(paths ...string) Option {
	return Method(http.MethodGet, paths...)
}

// Post sets the HTTP method to "POST"
func Post(paths ...string) Option {
	return Method(http.MethodPost, paths...)
}

// Put sets the HTTP method to "PUT"
func Put(paths ...string) Option {
	return Method(http.MethodPut, paths...)
}

// Patch sets the HTTP method to "PATCH"
func Patch(paths ...string) Option {
	return Method(http.MethodPatch, paths...)
}

// Delete sets the HTTP method to "DELETE"
func Delete(paths ...string) Option {
	return Method(http.MethodDelete, paths...)
}

// AddHeader adds a header value, using Header.Add()
func AddHeader(key, value string) Option {
	return OptionFunc(func(b *Requester) error {
		if b.Header == nil {
			b.Header = make(http.Header)
		}

		b.Header.Add(key, value)

		return nil
	})
}

// Header sets a header value, using Header.Set()
func Header(key, value string) Option {
	return OptionFunc(func(b *Requester) error {
		if b.Header == nil {
			b.Header = make(http.Header)
		}

		b.Header.Set(key, value)

		return nil
	})
}

// DeleteHeader deletes a header key, using Header.Del()
func DeleteHeader(key string) Option {
	return OptionFunc(func(b *Requester) error {
		b.Header.Del(key)

		return nil
	})
}

// BasicAuth sets the Authorization header to "Basic <encoded username and password>"
func BasicAuth(username, password string) Option {
	if username == "" && password == "" {
		return DeleteHeader(HeaderAuthorization)
	}

	return Header(HeaderAuthorization, BasicAuthHeader+basicAuth(username, password))
}

// basicAuth returns the base64 encoded username:password for basic auth copied from net/http
func basicAuth(username, password string) string {
	auth := username + ":" + password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// BearerAuth sets the Authorization header to "Bearer <token>"
func BearerAuth(token string) Option {
	if token == "" {
		return DeleteHeader(HeaderAuthorization)
	}

	return Header(HeaderAuthorization, BearerAuthHeader+token)
}

// URL sets the request URL
func URL(rawurl string) Option {
	return OptionFunc(func(b *Requester) error {
		u, err := url.Parse(rawurl)
		if err != nil {
			return fmt.Errorf("invalid url: %w", err)
		}

		b.URL = u

		return nil
	})
}

// RelativeURL resolves the arg as a relative URL references against the current URL, using the standard lib's url.URL.ResolveReference() method
func RelativeURL(paths ...string) Option {
	return OptionFunc(func(r *Requester) error {
		for _, p := range paths {
			u, err := url.Parse(p)
			if err != nil {
				return fmt.Errorf("invalid url: %w", err)
			}

			if r.URL == nil {
				r.URL = u
			} else {
				r.URL = r.URL.ResolveReference(u)
			}
		}

		return nil
	})
}

// AppendPath appends path elements to the end of the URL.Path
func AppendPath(elements ...string) Option {
	return OptionFunc(func(r *Requester) error {
		if len(elements) == 0 {
			return nil
		}

		var basePath string
		if r.URL != nil {
			basePath = r.URL.EscapedPath()
		}

		trailingSlash := strings.HasSuffix(basePath, "/")
		els := elements[:0]

		for _, e := range elements {
			trailingSlash = strings.HasSuffix(e, "/")
			e = strings.TrimFunc(e, func(r rune) bool {
				return unicode.IsSpace(r) || r == rune('/')
			})

			if len(e) > 0 {
				els = append(els, e)
			}
		}

		els = append(els, "")
		copy(els[1:], els)
		els[0] = strings.TrimSuffix(basePath, "/")

		newPath := strings.Join(els, "/")

		if trailingSlash {
			newPath += "/"
		}

		return RelativeURL(newPath).Apply(r)
	})
}

// QueryParams adds params to the Requester.QueryParams member
func QueryParams(queryStructs ...interface{}) Option {
	return OptionFunc(func(s *Requester) error {
		if s.QueryParams == nil {
			s.QueryParams = url.Values{}
		}

		for _, queryStruct := range queryStructs {
			var values url.Values

			switch t := queryStruct.(type) {
			case nil:
			case map[string]string:
				for key, value := range t {
					s.QueryParams.Add(key, value)
				}

				continue
			case map[string][]string:
				values = url.Values(t)
			case url.Values:
				values = t
			default:
				var err error

				values, err = goquery.Values(queryStruct)
				if err != nil {
					return fmt.Errorf("invalid query struct: %w", err)
				}
			}

			for key, values := range values {
				for _, value := range values {
					s.QueryParams.Add(key, value)
				}
			}
		}

		return nil
	})
}

// QueryParam adds a query parameter
func QueryParam(k, v string) Option {
	return OptionFunc(func(s *Requester) error {
		if k == "" {
			return nil
		}

		if s.QueryParams == nil {
			s.QueryParams = url.Values{}
		}

		s.QueryParams.Add(k, v)

		return nil
	})
}

// Body sets the body of the request
func Body(body interface{}) Option {
	return OptionFunc(func(b *Requester) error {
		b.Body = body

		return nil
	})
}

// WithMarshaler sets Requester.WithMarshaler
func WithMarshaler(m Marshaler) Option {
	return OptionFunc(func(b *Requester) error {
		b.Marshaler = m

		return nil
	})
}

// WithUnmarshaler sets Requester.WithUnmarshaler
func WithUnmarshaler(m Unmarshaler) Option {
	return OptionFunc(func(b *Requester) error {
		b.Unmarshaler = m

		return nil
	})
}

// Accept sets the Accept header
func Accept(accept string) Option {
	return Header(HeaderAccept, accept)
}

// ContentType sets the Content-Type header
func ContentType(contentType string) Option {
	return Header(HeaderContentType, contentType)
}

// Range sets the Range header
func Range(byteRange string) Option {
	return Header(HeaderRange, byteRange)
}

// Host sets Requester.Host
func Host(host string) Option {
	return OptionFunc(func(b *Requester) error {
		b.Host = host
		return nil
	})
}

func joinOpts(opts ...Option) Option {
	return OptionFunc(func(r *Requester) error {
		for _, opt := range opts {
			err := opt.Apply(r)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// JSON sets Requester.Marshaler to the JSONMarshaler
func JSON(indent bool) Option {
	return joinOpts(
		WithMarshaler(&JSONMarshaler{Indent: indent}),
		ContentType(ContentTypeJSON),
		Accept(ContentTypeJSON),
	)
}

// XML sets Requester.Marshaler to the XMLMarshaler
func XML(indent bool) Option {
	return joinOpts(
		WithMarshaler(&XMLMarshaler{Indent: indent}),
		ContentType(ContentTypeXML),
		Accept(ContentTypeXML),
	)
}

// Form sets Requester.Marshaler to the FormMarshaler which marshals the body into form-urlencoded
func Form() Option {
	return WithMarshaler(&FormMarshaler{})
}

// Client replaces Requester.Doer with an *http.Client
func Client(opts ...httpclient.Option) Option {
	return OptionFunc(func(b *Requester) error {
		c, err := httpclient.New(opts...)
		if err != nil {
			return err
		}

		b.Doer = c

		return nil
	})
}

// Use appends middleware to Requester.Middleware
func Use(m ...Middleware) Option {
	return OptionFunc(func(r *Requester) error {
		r.Middleware = append(r.Middleware, m...)

		return nil
	})
}

// WithDoer replaces Requester.Doer
func WithDoer(d Doer) Option {
	return OptionFunc(func(r *Requester) error {
		r.Doer = d

		return nil
	})
}
