package httptestutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"

	"github.com/felixge/httpsnoop"
)

// DumpTo wraps an http.Handler in a new handler
// the new handler dumps requests and responses to a writer, using the httputil.DumpRequest and
// httputil.DumpResponse functions
func DumpTo(handler http.Handler, writer io.Writer) http.Handler {
	// use the same default as http.Server
	if handler == nil {
		handler = http.DefaultServeMux
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			_, _ = fmt.Fprintf(writer, "error dumping request: %#v", err)
		} else {
			_, _ = writer.Write(append(dump, []byte("\r\n")...))
		}

		ex := Exchange{}

		w = httpsnoop.Wrap(w, hooks(&ex))

		handler.ServeHTTP(w, r)

		resp := http.Response{
			Proto:         r.Proto,
			ProtoMajor:    r.ProtoMajor,
			ProtoMinor:    r.ProtoMinor,
			StatusCode:    ex.StatusCode,
			Header:        w.Header(),
			Body:          io.NopCloser(bytes.NewReader(ex.ResponseBody.Bytes())),
			ContentLength: int64(ex.ResponseBody.Len()),
		}

		d, err := httputil.DumpResponse(&resp, true)
		if err != nil {
			fmt.Fprintf(writer, "error dumping response: %#v", err) // nolint: errcheck
		} else {
			writer.Write(append(d, []byte("\r\n")...)) // nolint: errcheck
		}
	})
}

// Dump writes requests and responses to the writer
func Dump(ts *httptest.Server, to io.Writer) {
	ts.Config.Handler = DumpTo(ts.Config.Handler, to)
}

// DumpToStdout writes requests and responses to os.Stdout
func DumpToStdout(ts *httptest.Server) {
	Dump(ts, os.Stdout)
}

type logFunc func(a ...interface{})

// Write implements io.Writer
func (f logFunc) Write(p []byte) (n int, err error) {
	f(string(p))

	return len(p), nil
}

// DumpToLog writes requests and responses to a logging function
func DumpToLog(ts *httptest.Server, logf func(a ...interface{})) {
	Dump(ts, logFunc(logf))
}
