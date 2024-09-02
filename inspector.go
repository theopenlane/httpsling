package httpsling

import (
	"bytes"
	"io"
	"net/http"
)

// Inspect installs and returns an Inspector
func Inspect(r *Requester) *Inspector {
	i := Inspector{}
	r.MustApply(&i)

	return &i
}

// Inspector is a Requester Option which captures requests and responses
type Inspector struct {
	// The last request sent by the client
	Request *http.Request
	// The last response received by the client
	Response *http.Response
	// The last client request body
	RequestBody *bytes.Buffer
	// The last client response body
	ResponseBody *bytes.Buffer
}

// Clear clears the inspector's fields
func (i *Inspector) Clear() {
	if i == nil {
		return
	}

	i.RequestBody = nil
	i.ResponseBody = nil
	i.Request = nil
	i.Response = nil
}

// Apply implements Option
func (i *Inspector) Apply(r *Requester) error {
	return r.Apply(Middleware(i.Wrap))
}

// Wrap implements Middleware
func (i *Inspector) Wrap(next Doer) Doer {
	return DoerFunc(func(req *http.Request) (*http.Response, error) {
		i.Request = req

		if req.Body != nil {
			reqBody, _ := io.ReadAll(req.Body)
			req.Body.Close()
			req.Body = io.NopCloser(bytes.NewReader(reqBody))
			i.RequestBody = bytes.NewBuffer(reqBody)
		}

		resp, err := next.Do(req)
		i.Response = resp

		if resp != nil && resp.Body != nil {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(respBody))
			i.ResponseBody = bytes.NewBuffer(respBody)
		}

		return resp, err
	})
}
