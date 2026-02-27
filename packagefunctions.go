package httpsling

import (
	"context"
	"io"
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
func Receive(into any, opts ...Option) (*http.Response, error) {
	return DefaultRequester.Receive(into, opts...)
}

// ReceiveWithContext does the same as Receive(), but attaches a Context to the request
func ReceiveWithContext(ctx context.Context, into any, opts ...Option) (*http.Response, error) {
	return DefaultRequester.ReceiveWithContext(ctx, into, opts...)
}

// ReceiveInto builds, sends and unmarshals into a typed value using the DefaultRequester.
func ReceiveInto[T any](opts ...Option) (*http.Response, T, error) {
	var out T
	resp, err := DefaultRequester.ReceiveWithContext(context.Background(), &out, opts...)

	return resp, out, err
}

// ReceiveIntoWithContext does the same as ReceiveInto but with a context.
func ReceiveIntoWithContext[T any](ctx context.Context, opts ...Option) (*http.Response, T, error) {
	var out T
	resp, err := DefaultRequester.ReceiveWithContext(ctx, &out, opts...)

	return resp, out, err
}

// ReceiveTo streams the response body into the writer using the DefaultRequester.
func ReceiveTo(w io.Writer, opts ...Option) (*http.Response, int64, error) {
	return DefaultRequester.ReceiveTo(context.Background(), w, opts...)
}

// ReceiveIntoWithError sends a request and decodes into S on success (2xx) or E on
// non-success. The error return wraps ErrUnsuccessfulResponse for non-2xx responses.
func ReceiveIntoWithError[S, E any](ctx context.Context, opts ...Option) (*http.Response, S, E, error) {
	var success S

	var failure E

	merged := make([]Option, len(opts)+1)
	copy(merged, opts)
	merged[len(opts)] = OnError(&failure)

	resp, err := DefaultRequester.ReceiveWithContext(ctx, &success, merged...)

	return resp, success, failure, err
}

// ReceiveToWithContext streams the response body into the writer with a context.
func ReceiveToWithContext(ctx context.Context, w io.Writer, opts ...Option) (*http.Response, int64, error) {
	return DefaultRequester.ReceiveTo(ctx, w, opts...)
}
