package httpsling

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspector(t *testing.T) {
	var dumpedReqBody []byte

	var doer DoerFunc = func(req *http.Request) (*http.Response, error) {
		dumpedReqBody, _ = io.ReadAll(req.Body)
		resp := &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(strings.NewReader("pong")),
		}

		return resp, nil
	}

	i := Inspector{}

	resp, body, err := Receive(&i, doer, Body("ping"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "pong", string(body))

	require.NotNil(t, i.Request)

	assert.Equal(t, "ping", i.RequestBody.String())
	assert.Equal(t, "ping", string(dumpedReqBody))

	require.NotNil(t, i.Response)
	assert.Equal(t, 201, i.Response.StatusCode)

	assert.Equal(t, "pong", i.ResponseBody.String())
}

func TestInspector_Clear(t *testing.T) {
	i := Inspector{
		Request:      &http.Request{},
		Response:     &http.Response{},
		RequestBody:  bytes.NewBuffer(nil),
		ResponseBody: bytes.NewBuffer(nil),
	}

	i.Clear()

	assert.Nil(t, i.Request)
	assert.Nil(t, i.Response)
	assert.Nil(t, i.RequestBody)
	assert.Nil(t, i.ResponseBody)

	assert.NotPanics(t, func() {
		(*Inspector)(nil).Clear()
	})
}

func TestInspect(t *testing.T) {
	r := MustNew()

	i := Inspect(r)

	_, _, err := r.Receive(MockDoer(201))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	assert.NotNil(t, i.Request)
	assert.Equal(t, 201, i.Response.StatusCode)
}

func ExampleInspect() {
	r := MustNew(
		MockDoer(201, Body("pong")),
		Header(HeaderAccept, ContentTypeText),
		Body("ping"),
	)

	i := Inspect(r)

	_, _, err := r.Receive(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(i.Request.Header.Get(HeaderAccept))
	fmt.Println(i.RequestBody.String())
	fmt.Println(i.Response.StatusCode)
	fmt.Println(i.ResponseBody.String())
}
