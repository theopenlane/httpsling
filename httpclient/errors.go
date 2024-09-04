package httpclient

import "errors"

var (
	// ErrMaxAttemptsExceeded is returned when the maximum number of attempts is exceeded
	ErrMaxAttemptsExceeded = errors.New("maximum number of attempts exceeded")
	// ErrInvalidTransportType is returned when the transport type is invalid
	ErrInvalidTransportType = errors.New("invalid transport type")
)
