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

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"color":"red"}`))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	b := &bytes.Buffer{}

	_, _, err := Receive(Get(ts.URL), Dump(b))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Log(b)

	assert.Contains(t, b.String(), "GET / HTTP/1.1")
	assert.Contains(t, b.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, b.String(), `{"color":"red"}`)
}

func TestDumpToLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"color":"red"}`))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	var args []interface{}

	_, _, err := Receive(Get(ts.URL), DumpToLog(func(a ...interface{}) {
		args = append(args, a...)
	}))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	assert.Len(t, args, 2)

	reqLog := args[0].(string)
	respLog := args[1].(string)

	assert.Contains(t, reqLog, "GET / HTTP/1.1")
	assert.Contains(t, respLog, "HTTP/1.1 200 OK")
	assert.Contains(t, respLog, `{"color":"red"}`)
}

func TestDumpToStout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"color":"red"}`))
		if err != nil {
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
		_, err := io.Copy(&buf, r)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		outC <- buf.String()
	}()

	_, _, err := Receive(Get(ts.URL), DumpToStout())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

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
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write([]byte(`{"color":"red"}`))
		if err != nil {
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
		_, err := io.Copy(&buf, r)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		outC <- buf.String()
	}()

	_, _, err := Receive(Get(ts.URL), DumpToStderr())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

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
		_, err := w.Write([]byte("boom!"))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}))

	defer ts.Close()

	r, err := New(Get(ts.URL))
	require.NoError(t, err)

	// without middleware
	resp, body, err := r.Receive(nil)
	require.NoError(t, err)
	require.Equal(t, 407, resp.StatusCode)
	require.Equal(t, "boom!", string(body))

	// add expect option
	r, err = r.With(ExpectCode(203))
	require.NoError(t, err)

	resp, body, err = r.Receive(nil)
	// body and response should still be returned
	assert.Equal(t, 407, resp.StatusCode)
	assert.Equal(t, "boom!", string(body))
	// but an error should be returned too
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected: 203")
	assert.Contains(t, err.Error(), "received: 407")
	assert.Equal(t, 407, merry.HTTPCode(err))

	// Using the option twice: latest option should win
	_, _, err = r.Receive(ExpectCode(407))
	require.NoError(t, err)

	// original requester's expect option should be unmodified
	_, _, err = r.Receive(nil)
	// but an error should be returned too
	require.Error(t, err)
	require.Equal(t, 407, merry.HTTPCode(err))
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

	resp, body, err = Receive(Get(ts.URL), ExpectSuccessCode())
	// body and response should still be returned
	assert.Equal(t, 407, resp.StatusCode)
	assert.Equal(t, "boom!", string(body))
	// but an error should be returned too
	require.Error(t, err)
	assert.Contains(t, err.Error(), "code: 407")
	assert.Equal(t, 407, merry.HTTPCode(err))

	// test positive path: if success code is returned, then no error should be returned
	successCodes := []int{200, 201, 204, 278}
	for _, code := range successCodes {
		codeToReturn = code
		_, _, err := Receive(Get(ts.URL), ExpectSuccessCode())
		require.NoError(t, err, "should not have received an error for code %v", code)
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

	_, err := Send(m)
	if err != nil {
		fmt.Println(err)
	}

	_, err = Send(Use(m))
	if err != nil {
		fmt.Println(err)
	}

	_ = Requester{
		Middleware: []Middleware{m},
	}
}

func ExampleDumpToLog() {
	_, err := Send(DumpToLog(func(a ...interface{}) {
		fmt.Println(a...)
	}))
	if err != nil {
		fmt.Println(err)
	}

	_, err = Send(DumpToLog(sdklog.Println))
	if err != nil {
		fmt.Println(err)
	}

	var t *testing.T

	_, err = Send(DumpToLog(t.Log))
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleExpectSuccessCode() {
	_, _, err := Receive(
		MockDoer(400),
		ExpectSuccessCode(),
	)

	fmt.Println(err.Error())
}

func ExampleExpectCode() {
	_, _, err := Receive(
		MockDoer(400),
		ExpectCode(201),
	)

	fmt.Println(err.Error())
}
