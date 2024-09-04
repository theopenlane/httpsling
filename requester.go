package httpsling

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
		return nil, err
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

// Clone returns a deep copy of a Requester
func (r *Requester) Clone() *Requester {
	s2 := *r
	s2.Header = r.Header.Clone()
	s2.Trailer = r.Trailer.Clone()
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
	requester, err := r.withOpts(opts...)
	if err != nil {
		return nil, err
	}

	bodyData, contentType, err := requester.getRequestBody()
	if err != nil {
		return nil, err
	}

	requestURL := ""
	if requester.URL != nil {
		requestURL = requester.URL.String()
	}

	req, err := http.NewRequestWithContext(ctx, requester.Method, requestURL, bodyData)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if requester.ContentLength != 0 {
		req.ContentLength = requester.ContentLength
	}

	if requester.GetBody != nil {
		req.GetBody = requester.GetBody
	}

	if requester.Host != "" {
		req.Host = requester.Host
	}

	req.TransferEncoding = requester.TransferEncoding
	req.Close = requester.Close

	if requester.Trailer != nil {
		req.Trailer = requester.Trailer.Clone()
	}

	if requester.Header != nil {
		req.Header = requester.Header.Clone()
	}

	// if we marshaled the body, use our content type
	if contentType != "" {
		req.Header.Set(HeaderContentType, contentType)
	}

	if len(requester.QueryParams) > 0 {
		req.URL.RawQuery = requester.getQueryParams(req)
	}

	return req, nil
}

func (r *Requester) getQueryParams(req *http.Request) string {
	if req.URL.RawQuery == "" {
		return r.QueryParams.Encode()
	}

	existingValues := req.URL.Query()

	for key, value := range r.QueryParams {
		for _, v := range value {
			existingValues.Add(key, v)
		}
	}

	return existingValues.Encode()
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
			return nil, "", fmt.Errorf("error marshaling body: %w", err)
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

// SendContext does the same as Send, but requires a context
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

// Do implements Doer
func (r *Requester) Do(req *http.Request) (*http.Response, error) {
	doer := r.Doer
	if doer == nil {
		doer = http.DefaultClient
	}

	resp, err := Wrap(doer, r.Middleware...).Do(req)

	return resp, err
}

// Receive creates a new HTTP request and returns the response
func (r *Requester) Receive(into interface{}, opts ...Option) (resp *http.Response, err error) {
	return r.ReceiveContext(context.Background(), into, opts...)
}

// ReceiveContext does the same as Receive, but requires a context
func (r *Requester) ReceiveContext(ctx context.Context, into interface{}, opts ...Option) (resp *http.Response, err error) {
	if opt, ok := into.(Option); ok {
		opts = append(opts, nil)
		copy(opts[1:], opts)
		opts[0] = opt
		into = nil
	}

	r, err = r.withOpts(opts...)
	if err != nil {
		return nil, err
	}

	resp, err = r.SendContext(ctx)

	body, bodyReadError := readBody(resp)

	if err != nil {
		return resp, err
	}

	if bodyReadError != nil {
		return resp, bodyReadError
	}

	if into != nil {
		unmarshaler := r.Unmarshaler
		if unmarshaler == nil {
			unmarshaler = DefaultUnmarshaler
		}

		err = unmarshaler.Unmarshal(body, resp.Header.Get(HeaderContentType), into)
	}

	return resp, err
}

func readBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil || resp.Body == http.NoBody {
		return nil, nil
	}

	defer resp.Body.Close()

	cls := resp.Header.Get(HeaderContentLength)

	var cl int64

	if cls != "" {
		cl, _ = strconv.ParseInt(cls, 10, 0)
	}

	buf := bytes.Buffer{}
	if cl > 0 {
		buf.Grow(int(cl))
	}

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
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

// HTTPClient returns the HTTP client used by the Requester
func (r *Requester) HTTPClient() *http.Client {
	client, ok := r.Doer.(*http.Client)
	if !ok {
		return nil
	}

	return client
}
