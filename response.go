package httpsling

import "net/http"

// IsSuccess checks if the response status code indicates success
func IsSuccess(resp *http.Response) bool {
	code := resp.StatusCode

	return code >= http.StatusOK && code <= http.StatusIMUsed
}
