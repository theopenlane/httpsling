package echoform

import (
	"fmt"
	"time"

	echo "github.com/theopenlane/echox"
)

// EchoContextAdapter acts as an adapter for an `echo.Context` object. It provides methods to interact with the underlying
// `echo.Context` object and extract information such as deadline, done channel, error, and values
// associated with specific keys from the context. The struct is used to enhance the functionality of
// the `echo.Context` object by providing additional methods and capabilities
type EchoContextAdapter struct {
	c echo.Context
}

// NewEchoContextAdapter takes echo.Context as a parameter and returns a pointer to
// a new EchoContextAdapter struct initialized with the provided echo.Context
func NewEchoContextAdapter(c echo.Context) *EchoContextAdapter {
	return &EchoContextAdapter{c: c}
}

// Deadline represents the time when the request should be completed
// deadline returns two values: deadline, which is the deadline time, and ok, indicating if a deadline is set or not
func (a *EchoContextAdapter) Deadline() (deadline time.Time, ok bool) {
	return a.c.Request().Context().Deadline()
}

// Done channel is used to receive a signal when the request context associated with the EchoContextAdapter is done or canceled
func (a *EchoContextAdapter) Done() <-chan struct{} {
	return a.c.Request().Context().Done()
}

// Err handles if an error occurred during the processing of the request
func (a *EchoContextAdapter) Err() error {
	return a.c.Request().Context().Err()
}

// Value implements the Value method of the context.Context interface
// used to retrieve a value associated with a specific key from the context
func (a *EchoContextAdapter) Value(key interface{}) interface{} {
	return a.c.Get(fmt.Sprintf("%v", key))
}
