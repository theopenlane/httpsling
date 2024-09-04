package httpsling

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	req, err := Request(Get("http://blue.com/red"))
	require.NoError(t, err)
	require.NotNil(t, req)
	require.Equal(t, "http://blue.com/red", req.URL.String())
}

type testContextKey string

const colorContextKey = testContextKey("color")

func TestRequestContext(t *testing.T) {
	req, err := RequestWithContext(
		context.WithValue(context.Background(), colorContextKey, "green"),
		Get("http://blue.com/red"),
	)
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "http://blue.com/red", req.URL.String())
	assert.Equal(t, "green", req.Context().Value(colorContextKey))
}

func TestSend(t *testing.T) {
	i := Inspector{}

	resp, err := Send(Get("/red"), WithDoer(MockDoer(204)), &i)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, 204, resp.StatusCode)
	assert.Equal(t, "/red", i.Request.URL.Path)
}

func TestSendContext(t *testing.T) {
	i := Inspector{}

	resp, err := SendWithContext(
		context.WithValue(context.Background(), colorContextKey, "blue"),
		Get("/profile"),
		WithDoer(MockDoer(204)),
		&i,
	)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, 204, resp.StatusCode)
	assert.Equal(t, "blue", i.Request.Context().Value(colorContextKey))
	assert.Equal(t, "/profile", i.Request.URL.Path)
}

func TestReceive(t *testing.T) {
	i := Inspector{}

	doer := MockDoer(205, Body(`{"count":25}`), JSON(false))

	var m testModel
	resp, err := Receive(&m, Get("/red"), WithDoer(doer), &i)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, 205, resp.StatusCode)
	assert.Equal(t, "/red", i.Request.URL.Path)
	assert.Equal(t, 25, m.Count)

	t.Run("Context", func(t *testing.T) {
		var m testModel

		i := Inspector{}

		resp, err := ReceiveWithContext(
			context.WithValue(context.Background(), colorContextKey, "yellow"),
			&m,
			Get("/red"),
			WithDoer(doer),
			&i,
		)
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 205, resp.StatusCode)
		assert.Equal(t, 25, m.Count)
		assert.Equal(t, "yellow", i.Request.Context().Value(colorContextKey))
		assert.Equal(t, "/red", i.Request.URL.Path)
	})
}

func ExampleRequest() {
	req, err := Request(Get("http://api.com/resource"))

	fmt.Println(req.URL.String(), err)

	// Output: http://api.com/resource <nil>
}
