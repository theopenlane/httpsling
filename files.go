package httpsling

import (
	"fmt"
	"net/http"
	"strings"
)

// Files is a map of form field names to a slice of files
type Files map[string][]File

// File represents a file that has been sent in an http request
type File struct {
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`
	// OriginalName is he name of the file from the client side / which was sent in the request
	OriginalName string `json:"original_name,omitempty"`
	// MimeType of the uploaded file
	MimeType string `json:"mime_type,omitempty"`
	// Size in bytes of the uploaded file
	Size int64 `json:"size,omitempty"`
}

// ValidationFunc is a type that can be used to dynamically validate a file
type ValidationFunc func(f File) error

// ErrResponseHandler is a custom error that should be used to handle errors when an upload fails
type ErrResponseHandler func(error) http.HandlerFunc

// NameGeneratorFunc allows you alter the name of the file before it is ultimately uploaded and stored
type NameGeneratorFunc func(s string) string

// FilesFromContext returns all files that have been uploaded during the request
func FilesFromContext(r *http.Request, key string) (Files, error) {
	files, ok := r.Context().Value(key).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files, nil
}

// FilesFromContextWithKey returns  all files that have been uploaded during the request
// and sorts by the provided form field
func FilesFromContextWithKey(r *http.Request, key string) ([]File, error) {
	files, ok := r.Context().Value(key).(Files)
	if !ok {
		return nil, ErrNoFilesUploaded
	}

	return files[key], nil
}

// MimeTypeValidator makes sure we only accept a valid mimetype.
// It takes in an array of supported mimes
func MimeTypeValidator(validMimeTypes ...string) ValidationFunc {
	return func(f File) error {
		for _, mimeType := range validMimeTypes {
			if strings.EqualFold(strings.ToLower(mimeType), f.MimeType) {
				return nil
			}
		}

		return fmt.Errorf("%w: %s", ErrUnsupportedMimeType, f.MimeType)
	}
}

// ChainValidators returns a validator that accepts multiple validating criteras
func ChainValidators(validators ...ValidationFunc) ValidationFunc {
	return func(f File) error {
		for _, validator := range validators {
			if err := validator(f); err != nil {
				return err
			}
		}

		return nil
	}
}
