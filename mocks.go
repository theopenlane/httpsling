package httpsling

import (
	"io"
	"net/http"
	"strings"
)

// MockDoer creates a Doer which returns a mocked response, for writing tests
func MockDoer(statusCode int, options ...Option) DoerFunc {
	return func(req *http.Request) (*http.Response, error) {
		resp := MockResponse(statusCode, options...)
		resp.Request = req

		return resp, nil
	}
}

// ChannelDoer returns a DoerFunc and a channel
func ChannelDoer() (chan<- *http.Response, DoerFunc) {
	input := make(chan *http.Response, 1)

	return input, func(req *http.Request) (*http.Response, error) {
		resp := <-input
		resp.Request = req

		return resp, nil
	}
}

// MockResponse creates an *http.Response from the Options
func MockResponse(statusCode int, options ...Option) *http.Response {
	r, err := Request(options...)
	if err != nil {
		panic(err)
	}

	resp := &http.Response{
		StatusCode:       statusCode,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header,
		Body:             r.Body,
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Trailer:          r.Trailer,
	}

	if resp.Body == nil {
		resp.Body = io.NopCloser(strings.NewReader(""))
	}

	return resp
}

// MockHandler returns an http.Handler which returns responses built from the args
func MockHandler(statusCode int, options ...Option) http.Handler {
	r := MustNew(options...)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		req, err := r.RequestWithContext(request.Context())
		if err != nil {
			panic(err)
		}

		h := writer.Header()
		for key, value := range req.Header {
			h[key] = value
		}

		writer.WriteHeader(statusCode)

		if req.Body != nil {
			if _, err := io.Copy(writer, req.Body); err != nil {
				panic(err)
			}
		}
	})
}

// ChannelHandler returns an http.Handler and an input channel
func ChannelHandler() (chan<- *http.Response, http.Handler) {
	input := make(chan *http.Response, 1)

	return input, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		resp := <-input

		h := writer.Header()
		for key, value := range resp.Header {
			h[key] = value
		}

		writer.WriteHeader(resp.StatusCode)

		if _, err := io.Copy(writer, resp.Body); err != nil {
			panic(err)
		}

		defer resp.Body.Close()
	})
}
