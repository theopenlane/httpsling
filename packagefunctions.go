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

// RequestWithContext does the same as Request(), but attaches a Context to the request
func RequestWithContext(ctx context.Context, opts ...Option) (*http.Request, error) {
	return DefaultRequester.RequestWithContext(ctx, opts...)
}

// Send uses the DefaultRequester to create a request and execute it
func Send(opts ...Option) (*http.Response, error) {
	return DefaultRequester.Send(opts...)
}

// SendWithContext does the same as Send(), but attaches a Context to the request
func SendWithContext(ctx context.Context, opts ...Option) (*http.Response, error) {
	return DefaultRequester.SendWithContext(ctx, opts...)
}

// Receive uses the DefaultRequester to create a request, execute it, and read the response
func Receive(into interface{}, opts ...Option) (*http.Response, error) {
	return DefaultRequester.Receive(into, opts...)
}

// ReceiveWithContext does the same as Receive(), but attaches a Context to the request
func ReceiveWithContext(ctx context.Context, into interface{}, opts ...Option) (*http.Response, error) {
	return DefaultRequester.ReceiveWithContext(ctx, into, opts...)
}
