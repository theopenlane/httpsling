package httptestutil

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/httpsling"
)

func TestDumpToStdout(t *testing.T) {
	ts := httptest.NewServer(httpsling.MockHandler(201,
		httpsling.Body("pong"),
		httpsling.JSON(true),
	))
	defer ts.Close()

	DumpToStdout(ts)

	_, err := Requester(ts).Receive(httpsling.Get("/test"), httpsling.Body("ping"))
	require.NoError(t, err)
}

func TestDump(t *testing.T) {
	ts := httptest.NewServer(httpsling.MockHandler(201,
		httpsling.Body(`{"ping":"pong"}`),
		httpsling.JSON(true),
	))
	defer ts.Close()

	buf := bytes.NewBuffer(nil)
	Dump(ts, buf)

	var out map[string]string
	resp, err := Requester(ts).Receive(&out, httpsling.Get("/test"), httpsling.Body("ping"))
	require.NoError(t, err)

	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "pong", out["ping"])
	require.NotEmpty(t, buf.Bytes())
	assert.Contains(t, buf.String(), "ping")
	assert.Contains(t, buf.String(), "pong")
}

func TestDumpToLog(t *testing.T) {
	ts := httptest.NewServer(httpsling.MockHandler(201,
		httpsling.Body(`{"ping":"pong"}`),
		httpsling.JSON(true),
	))
	defer ts.Close()

	DumpToLog(ts, t.Log)

	var out map[string]string
	Requester(ts).Receive(&out, httpsling.Get("/test"), httpsling.Body("ping"))
	require.Equal(t, "pong", out["ping"])
}

func TestDumpWithInspect(t *testing.T) {
	tests := []struct {
		name string
		f    func(*httptest.Server) (*bytes.Buffer, *Inspector)
	}{
		{"dumptheninspect", func(ts *httptest.Server) (*bytes.Buffer, *Inspector) {
			buf := bytes.Buffer{}
			Dump(ts, &buf)
			return &buf, Inspect(ts)
		}},
		{"inspectthendump", func(ts *httptest.Server) (*bytes.Buffer, *Inspector) {
			buf := bytes.Buffer{}
			i := Inspect(ts)
			Dump(ts, &buf)
			return &buf, i
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := httptest.NewServer(httpsling.MockHandler(201,
				httpsling.Body("pong"),
			))
			defer ts.Close()

			buf, i := test.f(ts)

			var out string
			resp, err := Requester(ts).Receive(&out, httpsling.Get("/test"), httpsling.Body("ping"))
			require.NoError(t, err)

			assert.Equal(t, 201, resp.StatusCode)
			assert.Equal(t, "pong", out)
			require.NotEmpty(t, buf.Bytes())
			assert.Contains(t, buf.String(), "ping")
			assert.Contains(t, buf.String(), "pong")

			ex := i.LastExchange()
			require.NotNil(t, ex)
			assert.Equal(t, 201, ex.StatusCode)
			assert.Equal(t, "ping", ex.RequestBody.String())
			assert.Equal(t, "pong", ex.ResponseBody.String())
		})
	}
}

func TestDumpToNilhandler(t *testing.T) {
	ts := httptest.NewServer(nil)
	defer ts.Close()

	var buf bytes.Buffer

	ts.Config.Handler = DumpTo(ts.Config.Handler, &buf)

	_, err := Requester(ts).Receive(nil)
	require.NoError(t, err)

	require.NotEmpty(t, buf)
}
