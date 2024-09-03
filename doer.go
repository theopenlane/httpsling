package httpsling

import "net/http"

// Doer executes http requests
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// DoerFunc adapts a function to implement Doer
type DoerFunc func(req *http.Request) (*http.Response, error)

// Apply implements the Option interface
func (f DoerFunc) Apply(r *Requester) error {
	r.Doer = f

	return nil
}

// Do implements the Doer interface
func (f DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// // Client returns a Doer that uses the given http.Client
// func (f DoerFunc) Client() *http.Client {
// 	var client *http.Client

// 	client = f.(*http.Client)

// 	return client

// }
