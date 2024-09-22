package httpsling

import (
	"errors"
)

var (
	// ErrUnsupportedContentType is returned when the content type is unsupported
	ErrUnsupportedContentType = errors.New("unsupported content type")
	// ErrUnsuccessfulResponse is returned when the response is unsuccessful
	ErrUnsuccessfulResponse = errors.New("unsuccessful response")
	// ErrNoFilesUploaded is returned when no files are found in a multipart form request
	ErrNoFilesUploaded = errors.New("no uploadable files found in request")
	// ErrUnsupportedMimeType is returned when the mime type is unsupported
	ErrUnsupportedMimeType = errors.New("unsupported mime type")
)
