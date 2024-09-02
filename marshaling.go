package httpsling

import (
	"encoding/json"
	"encoding/xml"
	"mime"
	"net/url"
	"strings"

	"github.com/ansel1/merry"
	goquery "github.com/google/go-querystring/query"
)

var DefaultMarshaler Marshaler = &JSONMarshaler{}

var DefaultUnmarshaler Unmarshaler = NewContentTypeUnmarshaler()

// Marshaler marshals values into a []byte
type Marshaler interface {
	Marshal(v interface{}) (data []byte, contentType string, err error)
}

// Unmarshaler unmarshals a []byte response body into a value
type Unmarshaler interface {
	Unmarshal(data []byte, contentType string, v interface{}) error
}

// MarshalFunc adapts a function to the Marshaler interface
type MarshalFunc func(v interface{}) ([]byte, string, error)

// Apply implements Option
func (f MarshalFunc) Apply(r *Requester) error {
	r.Marshaler = f
	return nil
}

// Marshal implements the Marshaler interface
func (f MarshalFunc) Marshal(v interface{}) ([]byte, string, error) {
	return f(v)
}

// UnmarshalFunc adapts a function to the Unmarshaler interface
type UnmarshalFunc func(data []byte, contentType string, v interface{}) error

// Apply implements Option
func (f UnmarshalFunc) Apply(r *Requester) error {
	r.Unmarshaler = f
	return nil
}

// Unmarshal implements the Unmarshaler interface
func (f UnmarshalFunc) Unmarshal(data []byte, contentType string, v interface{}) error {
	return f(data, contentType, v)
}

// JSONMarshaler implement Marshaler and Unmarshaler
type JSONMarshaler struct {
	Indent bool
}

// Unmarshal implements Unmarshaler
func (m *JSONMarshaler) Unmarshal(data []byte, _ string, v interface{}) error {
	return merry.Wrap(json.Unmarshal(data, v))
}

// Marshal implements Marshaler
func (m *JSONMarshaler) Marshal(v interface{}) (data []byte, contentType string, err error) {
	if m.Indent {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		data, err = json.Marshal(v)
	}

	return data, ContentTypeJSON, merry.Wrap(err)
}

// Apply implements Option
func (m *JSONMarshaler) Apply(r *Requester) error {
	r.Marshaler = m

	return nil
}

// XMLMarshaler implements Marshaler and Unmarshaler
type XMLMarshaler struct {
	Indent bool
}

// Unmarshal implements Unmarshaler
func (*XMLMarshaler) Unmarshal(data []byte, _ string, v interface{}) error {
	return merry.Wrap(xml.Unmarshal(data, v))
}

// Marshal implements Marshaler
func (m *XMLMarshaler) Marshal(v interface{}) (data []byte, contentType string, err error) {
	if m.Indent {
		data, err = xml.MarshalIndent(v, "", "  ")
	} else {
		data, err = xml.Marshal(v)
	}

	return data, ContentTypeXML, merry.Wrap(err)
}

// Apply implements Option
func (m *XMLMarshaler) Apply(r *Requester) error {
	r.Marshaler = m
	return nil
}

// FormMarshaler implements Marshaler
type FormMarshaler struct{}

// Marshal implements Marshaler
func (*FormMarshaler) Marshal(v interface{}) (data []byte, contentType string, err error) {
	switch t := v.(type) {
	case map[string][]string:
		urlV := url.Values(t)

		return []byte(urlV.Encode()), ContentTypeForm, nil
	case map[string]string:
		urlV := url.Values{}
		for key, value := range t {
			urlV.Set(key, value)
		}

		return []byte(urlV.Encode()), ContentTypeForm, nil
	case url.Values:
		return []byte(t.Encode()), ContentTypeForm, nil
	default:
		values, err := goquery.Values(v)
		if err != nil {
			return nil, "", merry.Prepend(err, "invalid form struct")
		}

		return []byte(values.Encode()), ContentTypeForm, nil
	}
}

// Apply implements Option
func (m *FormMarshaler) Apply(r *Requester) error {
	r.Marshaler = m

	return nil
}

// ContentTypeUnmarshaler selects an unmarshaler based on the content type
type ContentTypeUnmarshaler struct {
	Unmarshalers map[string]Unmarshaler
}

// NewContentTypeUnmarshaler returns a new ContentTypeUnmarshaler preconfigured to
// handle application/json and application/xml
func NewContentTypeUnmarshaler() *ContentTypeUnmarshaler {
	return &ContentTypeUnmarshaler{
		Unmarshalers: defaultUnmarshalers(),
	}
}

func defaultUnmarshalers() map[string]Unmarshaler {
	return map[string]Unmarshaler{
		ContentTypeJSON: &JSONMarshaler{},
		ContentTypeXML:  &XMLMarshaler{},
	}
}

// Unmarshal implements Unmarshaler
func (c *ContentTypeUnmarshaler) Unmarshal(data []byte, contentType string, v interface{}) error {
	if c.Unmarshalers == nil {
		c.Unmarshalers = map[string]Unmarshaler{
			ContentTypeJSON: &JSONMarshaler{},
			ContentTypeXML:  &XMLMarshaler{},
		}
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return merry.Prependf(err, "failed to parse content type: %s", contentType)
	}

	if u := c.Unmarshalers[mediaType]; u != nil {
		return u.Unmarshal(data, contentType, v)
	}

	if ct := generalMediaType(mediaType); ct != "" {
		if u := c.Unmarshalers[ct]; u != nil {
			return u.Unmarshal(data, contentType, v)
		}
	}

	return merry.Errorf("unsupported content type: %s", contentType)
}

// Apply implements Option
func (c *ContentTypeUnmarshaler) Apply(r *Requester) error {
	r.Unmarshaler = c
	return nil
}

// generalMediaType will return a media type with just the suffix as the subtype, e.g.
// application/vnd.api+json -> application/json
func generalMediaType(s string) string {
	i2 := strings.LastIndex(s, "+")
	if i2 > -1 && len(s) > i2+1 {
		i := strings.Index(s, "/")
		if i > -1 {
			return s[:i+1] + s[i2+1:]
		}
	}

	return ""
}

// MultiUnmarshaler is a legacy alias for ContentTypeUnmarshaler
type MultiUnmarshaler = ContentTypeUnmarshaler
