package httpsling

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockHandler(t *testing.T) {
	h := MockHandler(201,
		JSON(false),
		Body(map[string]interface{}{"color": "blue"}),
	)

	ts := httptest.NewServer(h)
	defer ts.Close()

	resp, body, err := Receive(Get(ts.URL))
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)
	assert.JSONEq(t, `{"color":"blue"}`, string(body))
	assert.Contains(t, resp.Header.Get(HeaderContentType), ContentTypeJSON)
}

func TestChannelHandler(t *testing.T) {
	in, h := ChannelHandler() // nolint: bodyclose

	ts := httptest.NewServer(h)
	defer ts.Close()

	in <- MockResponse(201, JSON(false), // nolint: bodyclose
		Body(map[string]interface{}{"color": "blue"}))

	resp, body, err := Receive(Get(ts.URL))
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)
	assert.JSONEq(t, `{"color":"blue"}`, string(body))
	assert.Contains(t, resp.Header.Get(HeaderContentType), ContentTypeJSON)
}

func TestMockResponse(t *testing.T) {
	resp := MockResponse(201,
		JSON(false),
		Body(map[string]interface{}{"color": "red"}),
	)

	defer resp.Body.Close()

	require.NotNil(t, resp)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Contains(t, resp.Header.Get(HeaderContentType), ContentTypeJSON)

	b, _ := io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"color":"red"}`, string(b))

	resp = MockResponse(500)
	assert.NotNil(t, resp.Body)

	defer resp.Body.Close()
}

func TestMockDoer(t *testing.T) {
	d := MockDoer(201,
		JSON(false),
		Body(map[string]interface{}{"color": "blue"}),
	)

	req, err := Request(Get("/profile"), d)
	require.NoError(t, err)

	resp, err := d.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.NotNil(t, resp)

	assert.Equal(t, req, resp.Request)
	assert.Equal(t, 201, resp.StatusCode)

	assert.Contains(t, resp.Header.Get(HeaderContentType), ContentTypeJSON)

	b, _ := io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"color":"blue"}`, string(b))
}

func TestChannelDoer(t *testing.T) {
	in, d := ChannelDoer() // nolint: bodyclose

	in <- MockResponse(201, // nolint: bodyclose
		JSON(false),
		Body(map[string]interface{}{"color": "blue"}),
	)

	req, err := Request(Get("/profile"), d)
	require.NoError(t, err)

	resp, err := d.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.NotNil(t, resp)

	assert.Equal(t, req, resp.Request)
	assert.Equal(t, 201, resp.StatusCode)

	assert.Contains(t, resp.Header.Get(HeaderContentType), ContentTypeJSON)

	b, _ := io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"color":"blue"}`, string(b))
}

func ExampleMockDoer() {
	d := MockDoer(201,
		JSON(false),
		Body(map[string]interface{}{"color": "blue"}),
	)

	// Since DoerFunc is an Option, it can be passed directly to functions
	// which accept Options.
	resp, body, _ := Receive(d)
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get(HeaderContentType))
	fmt.Println(string(body))

	// Output:
	// 201
	// application/json;charset=utf-8
	// {"color":"blue"}
}

func ExampleMockHandler() {
	h := MockHandler(201,
		JSON(false),
		Body(map[string]interface{}{"color": "blue"}),
	)

	ts := httptest.NewServer(h)
	defer ts.Close()

	resp, body, _ := Receive(URL(ts.URL))

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get(HeaderContentType))
	fmt.Println(string(body))

	// Output:
	// 201
	// application/json;charset=utf-8
	// {"color":"blue"}
}

func ExampleChannelDoer() {
	in, d := ChannelDoer() // nolint: bodyclose

	in <- &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(strings.NewReader("pong")),
	}

	resp, body, _ := Receive(d)

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 201
	// pong
}

func ExampleChannelHandler() {
	in, h := ChannelHandler() // nolint: bodyclose

	ts := httptest.NewServer(h)
	defer ts.Close()

	in <- &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(strings.NewReader("pong")),
	}

	resp, body, _ := Receive(URL(ts.URL))

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 201
	// pong
}
