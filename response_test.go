package httpsling_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/httpsling"
)

func TestIsSuccess(t *testing.T) {
	testCases := []struct {
		name     string
		code     int
		expected bool
	}{
		{
			name:     "OK",
			code:     http.StatusOK,
			expected: true,
		},
		{
			name:     "Unauthorized",
			code:     http.StatusUnauthorized,
			expected: false,
		},
		{
			name:     "Created",
			code:     http.StatusCreated,
			expected: true,
		},
		{
			name:     "InternalServerError",
			code:     http.StatusInternalServerError,
			expected: false,
		},
		{
			name:     "BadRequest",
			code:     http.StatusBadRequest,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.code,
			}

			result := httpsling.IsSuccess(resp)

			assert.Equal(t, tc.expected, result)
		})
	}
}
