package httpsling

import (
	"errors"
)

var (
	// ErrUnsupportedContentType is returned when the content type is unsupported
	ErrUnsupportedContentType = errors.New("unsupported content type")
	// ErrUnsuccessfulResponse is returned when the response is unsuccessful
	ErrUnsuccessfulResponse = errors.New("unsuccessful response")
)
