package httpsling

import (
	"bytes"
	"fmt"
	"io"
	sdklog "log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)

		if _, err := w.Write([]byte(`{"color":"red"}`)); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	b := &bytes.Buffer{}

	resp, _, err := Receive(Get(ts.URL), Dump(b))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	defer resp.Body.Close()

	t.Log(b)

	assert.Contains(t, b.String(), "GET / HTTP/1.1")
	assert.Contains(t, b.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, b.String(), `{"color":"red"}`)
}

func TestDumpToLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)

		if _, err := w.Write([]byte(`{"color":"red"}`)); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	var args []interface{}

	resp, _, err := Receive(Get(ts.URL), DumpToLog(func(a ...interface{}) {
		args = append(args, a...)
	}))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	defer resp.Body.Close()

	assert.Len(t, args, 2)

	reqLog := args[0].(string)
	respLog := args[1].(string)

	assert.Contains(t, reqLog, "GET / HTTP/1.1")
	assert.Contains(t, respLog, "HTTP/1.1 200 OK")
	assert.Contains(t, respLog, `{"color":"red"}`)
}

func TestDumpToStout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)

		if _, err := w.Write([]byte(`{"color":"red"}`)); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() { os.Stdout = old }()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer

		if _, err := io.Copy(&buf, r); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		outC <- buf.String()
	}()

	resp, _, err := Receive(Get(ts.URL), DumpToStout())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	defer resp.Body.Close()

	// back to normal state
	os.Stdout = old // restoring the real stdout

	w.Close()

	out := <-outC

	assert.Contains(t, out, "GET / HTTP/1.1")
	assert.Contains(t, out, "HTTP/1.1 200 OK")
	assert.Contains(t, out, `{"color":"red"}`)
}

func TestDumpToSterr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)

		if _, err := w.Write([]byte(`{"color":"red"}`)); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	old := os.Stderr // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stderr = w

	defer func() { os.Stderr = old }()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer

		if _, err := io.Copy(&buf, r); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		outC <- buf.String()
	}()

	resp, _, err := Receive(Get(ts.URL), DumpToStderr())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	defer resp.Body.Close()

	// back to normal state
	os.Stderr = old // restoring the real stdout

	w.Close()

	out := <-outC

	assert.Contains(t, out, "GET / HTTP/1.1")
	assert.Contains(t, out, "HTTP/1.1 200 OK")
	assert.Contains(t, out, `{"color":"red"}`)
}

func TestExpectCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(407)

		if _, err := w.Write([]byte("boom!")); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	r, err := New(Get(ts.URL))
	require.NoError(t, err)

	// without middleware
	resp, body, err := r.Receive(nil)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, 407, resp.StatusCode)
	require.Equal(t, "boom!", string(body))

	// add expect option
	r, err = r.With(ExpectCode(203))
	require.NoError(t, err)

	resp, body, err = r.Receive(nil)

	// but an error should be returned too
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected: 203")
	assert.Contains(t, err.Error(), "received: 407")

	defer resp.Body.Close()

	// body and response should still be returned
	assert.Equal(t, 407, resp.StatusCode)
	assert.Equal(t, "boom!", string(body))

	// Using the option twice: latest option should win
	resp, _, err = r.Receive(ExpectCode(407))
	require.NoError(t, err)

	defer resp.Body.Close()

	// original requester's expect option should be unmodified
	resp, _, err = r.Receive(nil)
	// but an error should be returned too
	require.Error(t, err)

	defer resp.Body.Close()
}

func TestExpectSuccessCode(t *testing.T) {
	codeToReturn := 407
	bodyToReturn := "boom!"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(codeToReturn)
		_, _ = w.Write([]byte(bodyToReturn))
	}))

	defer ts.Close()

	// without middleware
	resp, body, err := Receive(Get(ts.URL))
	require.NoError(t, err)
	require.Equal(t, 407, resp.StatusCode)
	require.Equal(t, "boom!", string(body))

	defer resp.Body.Close()

	resp, body, err = Receive(Get(ts.URL), ExpectSuccessCode())
	// body and response should still be returned
	assert.Equal(t, 407, resp.StatusCode)
	assert.Equal(t, "boom!", string(body))
	// but an error should be returned too
	require.Error(t, err)
	assert.Contains(t, err.Error(), "code: 407")

	defer resp.Body.Close()

	// test positive path: if success code is returned, then no error should be returned
	successCodes := []int{200, 201, 204, 278}
	for _, code := range successCodes {
		codeToReturn = code
		resp, _, err := Receive(Get(ts.URL), ExpectSuccessCode())
		require.NoError(t, err, "should not have received an error for code %v", code)

		defer resp.Body.Close()
	}
}

func ExampleMiddleware() {
	var m Middleware = func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			d, _ := httputil.DumpRequest(req, true)
			fmt.Println(string(d))
			return next.Do(req)
		})
	}

	resp, err := Send(m)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	resp, err = Send(Use(m))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	_ = Requester{
		Middleware: []Middleware{m},
	}
}

func ExampleDumpToLog() {
	resp, err := Send(DumpToLog(func(a ...interface{}) {
		fmt.Println(a...)
	}))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	resp, err = Send(DumpToLog(sdklog.Println))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	var t *testing.T

	resp, err = Send(DumpToLog(t.Log))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
}

func ExampleExpectSuccessCode() {
	resp, _, err := Receive(
		MockDoer(400),
		ExpectSuccessCode(),
	)

	fmt.Println(err.Error())

	defer resp.Body.Close()
}

func ExampleExpectCode() {
	resp, _, err := Receive(
		MockDoer(400),
		ExpectCode(201),
	)

	fmt.Println(err.Error())

	defer resp.Body.Close()
}
