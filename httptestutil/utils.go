package httptestutil

import (
	"net/http/httptest"

	"github.com/theopenlane/httpsling"
)

// Requester creates a Requester instance which is pre-configured to send requests to the test server
func Requester(ts *httptest.Server, options ...httpsling.Option) *httpsling.Requester {
	r := httpsling.MustNew(httpsling.URL(ts.URL), httpsling.WithDoer(ts.Client()))
	r.MustApply(options...)

	return r
}

// Inspect installs and returns an Inspector; the Inspector captures exchanges with the test server
func Inspect(ts *httptest.Server) *Inspector {
	i := NewInspector(0)
	ts.Config.Handler = i.Wrap(ts.Config.Handler)

	return i
}
