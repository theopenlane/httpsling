package httpsling

import (
	"context"
	"net/http"
)

var DefaultRequester = Requester{}

// Request uses the DefaultRequester to create a request
func Request(opts ...Option) (*http.Request, error) {
	return DefaultRequester.Request(opts...)
}

// RequestContext does the same as Request(), but attaches a Context to the request
func RequestContext(ctx context.Context, opts ...Option) (*http.Request, error) {
	return DefaultRequester.RequestContext(ctx, opts...)
}

// Send uses the DefaultRequester to create a request and execute it
func Send(opts ...Option) (*http.Response, error) {
	return DefaultRequester.Send(opts...)
}

// SendContext does the same as Send(), but attaches a Context to the request
func SendContext(ctx context.Context, opts ...Option) (*http.Response, error) {
	return DefaultRequester.SendContext(ctx, opts...)
}

// ReceiveContext does the same as Receive(), but attaches a Context to the request
func ReceiveContext(ctx context.Context, into interface{}, opts ...Option) (*http.Response, []byte, error) {
	return DefaultRequester.ReceiveContext(ctx, into, opts...)
}

// Receive uses the DefaultRequester to create a request, execute it, and read the response
func Receive(into interface{}, opts ...Option) (*http.Response, []byte, error) {
	return DefaultRequester.Receive(into, opts...)
}
