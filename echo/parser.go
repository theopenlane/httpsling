package echoform

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/mazrean/formstream"
	echo "github.com/theopenlane/echox"
)

type FormParser struct {
	*formstream.Parser
	reader io.Reader
}

// NewFormParser creates a new multipart form parser
func NewFormParser(c echo.Context, options ...formstream.ParserOption) (*FormParser, error) {
	contentType := c.Request().Header.Get("Content-Type")

	d, params, err := mime.ParseMediaType(contentType)
	if err != nil || d != "multipart/form-data" {
		return nil, http.ErrNotMultipart
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}

	return &FormParser{
		Parser: formstream.NewParser(boundary, options...),
		reader: c.Request().Body,
	}, nil
}

// Parse parses the request body; it returns the echo.HTTPError if the hook function returns an echo.HTTPError
func (p *FormParser) Parse() error {
	err := p.Parser.Parse(p.reader)

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	return err
}
