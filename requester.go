package httpsling

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ansel1/merry"
)

// Requester is a struct that contains the information needed to make an HTTP request
type Requester struct {
	// Method is the HTTP method to use for the request
	Method string
	// URL is the URL to request
	URL *url.URL
	// Header supplies the request headers; if the Content-Type header is set here, it will override the Content-Type header supplied by the Marshaler
	Header http.Header
	// GetBody is a function that returns a ReadCloser for the request body
	GetBody func() (io.ReadCloser, error)
	// ContentLength is the length of the request body
	ContentLength int64
	// TransferEncoding is the transfer encoding for the request body
	TransferEncoding []string
	// Close indicates whether the connection should be closed after the request
	Close bool
	// Host is the host to use for the request
	Host string
	// Trailer is the trailer for the request
	Trailer http.Header
	// QueryParams are added to the request, in addition to any query params already encoded in the URL
	QueryParams url.Values
	// Body can be set to a string, []byte, io.Reader, or a struct; if set to a string, []byte, or io.Reader, the value will be used as the body of the request
	// If set to a struct, the Marshaler will be used to marshal the value into the request body
	Body interface{}
	// Marshaler will be used to marshal the Body value into the body of the request.
	Marshaler Marshaler
	// Doer holds the HTTP client for used to execute httpsling
	Doer Doer
	// Middleware wraps the Doer
	Middleware []Middleware
	// Unmarshaler will be used by the Receive methods to unmarshal the response body
	Unmarshaler Unmarshaler
}

// New returns a new Requester, applying all options
func New(options ...Option) (*Requester, error) {
	b := &Requester{}
	err := b.Apply(options...)

	if err != nil {
		return nil, merry.Wrap(err)
	}

	return b, nil
}

// MustNew creates a new Requester, applying all options
func MustNew(options ...Option) *Requester {
	b := &Requester{}
	b.MustApply(options...)

	return b
}

func cloneURL(url *url.URL) *url.URL {
	if url == nil {
		return nil
	}

	urlCopy := *url

	return &urlCopy
}

func cloneValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}

	v2 := make(url.Values, len(v))

	for key, value := range v {
		v2[key] = value
	}

	return v2
}

func cloneHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}

	h2 := make(http.Header)

	for key, value := range h {
		h2[key] = value
	}

	return h2
}

// Clone returns a deep copy of a Requester
func (r *Requester) Clone() *Requester {
	s2 := *r
	s2.Header = cloneHeader(r.Header)
	s2.Trailer = cloneHeader(r.Trailer)
	s2.URL = cloneURL(r.URL)
	s2.QueryParams = cloneValues(r.QueryParams)
	return &s2
}

// Request returns a new http.Request
func (r *Requester) Request(opts ...Option) (*http.Request, error) {
	return r.RequestContext(context.Background(), opts...)
}

// RequestContext does the same as Request, but requires a context
func (r *Requester) RequestContext(ctx context.Context, opts ...Option) (*http.Request, error) {
	reqs, err := r.withOpts(opts...)
	if err != nil {
		return nil, err
	}

	// marshal body, if applicable
	bodyData, ct, err := reqs.getRequestBody()
	if err != nil {
		return nil, err
	}

	urlS := ""
	if reqs.URL != nil {
		urlS = reqs.URL.String()
	}

	req, err := http.NewRequest(reqs.Method, urlS, bodyData)
	if err != nil {
		return nil, merry.Prepend(err, "creating request")
	}

	// if we marshaled the body, use our content type
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}

	if reqs.ContentLength != 0 {
		req.ContentLength = reqs.ContentLength
	}

	if reqs.GetBody != nil {
		req.GetBody = reqs.GetBody
	}

	// copy the host
	if reqs.Host != "" {
		req.Host = reqs.Host
	}

	req.TransferEncoding = reqs.TransferEncoding
	req.Close = reqs.Close
	req.Trailer = reqs.Trailer

	// copy Headers pairs into new Header map
	for k, v := range reqs.Header {
		req.Header[k] = v
	}

	if len(reqs.QueryParams) > 0 {
		if req.URL.RawQuery != "" {
			existingValues := req.URL.Query()

			for key, value := range reqs.QueryParams {
				for _, v := range value {
					existingValues.Add(key, v)
				}
			}
			req.URL.RawQuery = existingValues.Encode()
		} else {
			req.URL.RawQuery = reqs.QueryParams.Encode()
		}
	}

	return req.WithContext(ctx), nil
}

// getRequestBody returns the io.Reader which should be used as the body of new Requester
func (r *Requester) getRequestBody() (body io.Reader, contentType string, _ error) {
	switch v := r.Body.(type) {
	case nil:
		return nil, "", nil
	case io.Reader:
		return v, "", nil
	case string:
		return strings.NewReader(v), "", nil
	case []byte:
		return bytes.NewReader(v), "", nil
	default:
		marshaler := r.Marshaler
		if marshaler == nil {
			marshaler = DefaultMarshaler
		}
		b, ct, err := marshaler.Marshal(r.Body)

		if err != nil {
			return nil, "", merry.Prepend(err, "marshaling body")
		}

		return bytes.NewReader(b), ct, nil
	}
}

// Send executes a request with the Doer
func (r *Requester) Send(opts ...Option) (*http.Response, error) {
	return r.SendContext(context.Background(), opts...)
}

// withOpts is like With(), but skips the clone if there are no options to apply
func (r *Requester) withOpts(opts ...Option) (*Requester, error) {
	if len(opts) > 0 {
		return r.With(opts...)
	}

	return r, nil
}

// SendContext does the same as Request, but requires a context
func (r *Requester) SendContext(ctx context.Context, opts ...Option) (*http.Response, error) {
	reqs, err := r.withOpts(opts...)
	if err != nil {
		return nil, err
	}

	req, err := reqs.RequestContext(ctx)
	if err != nil {
		return nil, err
	}

	return reqs.Do(req)
}

// Do implements Doer.
func (r *Requester) Do(req *http.Request) (*http.Response, error) {
	doer := r.Doer
	if doer == nil {
		doer = http.DefaultClient
	}

	resp, err := Wrap(doer, r.Middleware...).Do(req)

	return resp, merry.Wrap(err)
}

// Receive creates a new HTTP request and returns the response
func (r *Requester) Receive(into interface{}, opts ...Option) (resp *http.Response, body []byte, err error) {
	return r.ReceiveContext(context.Background(), into, opts...)
}

// ReceiveContext does the same as Receive, but requires a context
func (r *Requester) ReceiveContext(ctx context.Context, into interface{}, opts ...Option) (resp *http.Response, body []byte, err error) {
	if opt, ok := into.(Option); ok {
		opts = append(opts, nil)
		copy(opts[1:], opts)
		opts[0] = opt
		into = nil
	}

	r, err = r.withOpts(opts...)
	if err != nil {
		return nil, nil, err
	}

	resp, err = r.SendContext(ctx)

	body, bodyReadError := readBody(resp)

	if err != nil {
		return resp, body, err
	}

	if bodyReadError != nil {
		return resp, body, bodyReadError
	}

	if into != nil {
		unmarshaler := r.Unmarshaler
		if unmarshaler == nil {
			unmarshaler = DefaultUnmarshaler
		}

		err = unmarshaler.Unmarshal(body, resp.Header.Get("Content-Type"), into)
	}

	return resp, body, err
}

func readBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil || resp.Body == http.NoBody {
		return nil, nil
	}

	defer resp.Body.Close()

	cls := resp.Header.Get("Content-Length")

	var cl int64

	if cls != "" {
		cl, _ = strconv.ParseInt(cls, 10, 0)
	}

	buf := bytes.Buffer{}
	if cl > 0 {
		buf.Grow(int(cl))
	}

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, merry.Prepend(err, "reading response body")
	}

	return buf.Bytes(), nil
}

// Params returns the QueryParams
func (r *Requester) Params() url.Values {
	if r.QueryParams == nil {
		r.QueryParams = url.Values{}
	}

	return r.QueryParams
}

// Headers returns the Header
func (r *Requester) Headers() http.Header {
	if r.Header == nil {
		r.Header = http.Header{}
	}

	return r.Header
}

// Trailers returns the Trailer, initializing it if necessary
func (r *Requester) Trailers() http.Header {
	if r.Trailer == nil {
		r.Trailer = http.Header{}
	}

	return r.Trailer
}
