package httptestutil

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/httpsling"
)

func TestNewInspector(t *testing.T) {
	i := NewInspector(0)

	ts := httptest.NewServer(httpsling.MockHandler(201, httpsling.Body("pong")))
	defer ts.Close()

	origHandler := ts.Config.Handler

	ts.Config.Handler = i.Wrap(origHandler)

	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))

	assert.Len(t, i.Exchanges, 2)

	i = NewInspector(5)

	ts.Config.Handler = i.Wrap(origHandler)

	// run ten requests
	for i := 0; i < 10; i++ {
		Requester(ts).Receive(httpsling.Get("/test"))
	}

	// channel should only have buffered 5
	assert.Len(t, i.Exchanges, 5)
}

func TestInspector(t *testing.T) {
	ts := httptest.NewServer(httpsling.MockHandler(201, httpsling.Body("pong")))
	defer ts.Close()

	is := Inspect(ts)

	var out string
	resp, err := Requester(ts).Receive(&out, httpsling.Get("/test"), httpsling.Body("ping"))
	require.NoError(t, err)

	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "pong", out)

	ex := is.LastExchange()
	require.NotNil(t, ex)
	assert.Equal(t, "/test", ex.Request.URL.Path)
	assert.Equal(t, "ping", ex.RequestBody.String())
	assert.Equal(t, 201, ex.StatusCode)
	assert.Equal(t, "pong", ex.ResponseBody.String())
}

func TestInspectorNextExchange(t *testing.T) {
	var count int

	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(201)
		writer.Write([]byte("pong" + strconv.Itoa(count)))
		count++
	}))
	defer ts.Close()

	is := Inspect(ts)

	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))

	var exchanges []*Exchange

	for {
		ex := is.NextExchange()
		if ex == nil {
			break
		}
		exchanges = append(exchanges, ex)
	}

	assert.Len(t, exchanges, 3)
	assert.Equal(t, "pong0", exchanges[0].ResponseBody.String())
	assert.Equal(t, "pong1", exchanges[1].ResponseBody.String())
	assert.Equal(t, "pong2", exchanges[2].ResponseBody.String())
}

func TestInspectorLastExchange(t *testing.T) {
	ts := httptest.NewServer(nil)
	defer ts.Close()

	var count int
	ts.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(201)
		writer.Write([]byte("pong" + strconv.Itoa(count)))
		count++
	})

	is := Inspect(ts)

	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))

	ex := is.LastExchange()

	require.NotNil(t, ex)
	assert.Equal(t, "pong2", ex.ResponseBody.String())

	require.Nil(t, is.LastExchange())
}

func TestInspectorDrain(t *testing.T) {
	ts := httptest.NewServer(nil)
	defer ts.Close()

	var count int
	ts.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(201)
		writer.Write([]byte("pong" + strconv.Itoa(count)))
		count++
	})

	is := Inspect(ts)

	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))

	drain := is.Drain()

	require.Len(t, drain, 3)
	assert.Equal(t, "pong1", drain[1].ResponseBody.String())
	require.Nil(t, is.LastExchange())
}

func TestInspectorClear(t *testing.T) {
	ts := httptest.NewServer(nil)
	defer ts.Close()

	var count int
	ts.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(201)
		writer.Write([]byte("pong" + strconv.Itoa(count)))
		count++
	})

	is := Inspect(ts)

	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))
	Requester(ts).Receive(httpsling.Get("/test"))

	require.Len(t, is.Exchanges, 3)

	is.Clear()

	require.Empty(t, is.Exchanges)

	t.Run("nil", func(t *testing.T) {
		var i *Inspector
		assert.NotPanics(t, func() {
			i.Clear()
		})
	})
}

func TestInspectorReadFrom(t *testing.T) {
	// fixed a bug in the hook func's ReadFrom hook.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeTextUTF8)
		w.WriteHeader(201)
		readerFrom := w.(io.ReaderFrom)
		readerFrom.ReadFrom(strings.NewReader("pong"))
		readerFrom.ReadFrom(strings.NewReader("kilroy"))
	}))
	defer ts.Close()

	i := Inspect(ts)

	var out string
	Requester(ts).Receive(&out, httpsling.Get("/test"), httpsling.Body("ping"))
	assert.Equal(t, "pongkilroy", out)
	assert.Equal(t, "pongkilroy", i.LastExchange().ResponseBody.String())
}

func TestInspectNilhandler(t *testing.T) {
	ts := httptest.NewServer(nil)
	defer ts.Close()

	i := Inspect(ts)

	_, err := Requester(ts).Receive(nil)
	require.NoError(t, err)

	require.NotNil(t, i.LastExchange())
}

func ExampleInspector_NextExchange() {
	i := NewInspector(0)

	var h http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`pong`))
	})

	h = i.Wrap(h)

	ts := httptest.NewServer(h)
	defer ts.Close()

	httpsling.Receive(httpsling.Get(ts.URL), httpsling.Body("ping1"))
	httpsling.Receive(httpsling.Get(ts.URL), httpsling.Body("ping2"))

	fmt.Println(i.NextExchange().RequestBody.String())
	fmt.Println(i.NextExchange().RequestBody.String())
	fmt.Println(i.NextExchange())

	// Output:
	// ping1
	// ping2
	// <nil>
}

func ExampleInspector_LastExchange() {
	i := NewInspector(0)

	var h http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`pong`))
	})

	h = i.Wrap(h)

	ts := httptest.NewServer(h)
	defer ts.Close()

	httpsling.Receive(httpsling.Get(ts.URL), httpsling.Body("ping1"))
	httpsling.Receive(httpsling.Get(ts.URL), httpsling.Body("ping2"))

	fmt.Println(i.LastExchange().RequestBody.String())
	fmt.Println(i.LastExchange())

	// Output:
	// ping2
	// <nil>
}
