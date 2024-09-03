package httpsling

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FakeModel struct {
	Text          string  `json:"text,omitempty"`
	FavoriteCount int64   `json:"favorite_count,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

var modelA = FakeModel{Text: "note", FavoriteCount: 12}

func failOption() OptionFunc {
	return func(_ *Requester) error {
		return errors.New("boom") // nolint: err113
	}
}

func TestNew(t *testing.T) {
	reqs, err := New(URL("green"), Method("POST"))
	require.NoError(t, err)
	require.NotNil(t, reqs)

	// options were applied
	require.Equal(t, "green", reqs.URL.String())
	require.Equal(t, "POST", reqs.Method)

	t.Run("error", func(t *testing.T) {
		_, err := New(failOption())
		require.Error(t, err)

		require.ErrorContains(t, err, "boom")
	})
}

func TestMustNew(t *testing.T) {
	reqs := MustNew(URL("green"), Method("POST"))
	require.NotNil(t, reqs)
	// options were applied
	require.Equal(t, "green", reqs.URL.String())
	require.Equal(t, "POST", reqs.Method)

	require.Panics(t, func() {
		MustNew(failOption())
	})
}

func TestRequester_Clone(t *testing.T) {
	cases := [][]Option{
		{Get(), URL("http: //example.com")},
		{URL("http://example.com")},
		{QueryParams(url.Values{})},
		{QueryParams(paramsA)},
		{QueryParams(paramsA, paramsB)},
		{Body(&FakeModel{Text: "a"})},
		{Body(FakeModel{Text: "a"})},
		{Header("Content-Type", "application/json")},
		{AddHeader("A", "B"), AddHeader("a", "c")},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			reqs, err := New(c...)
			require.NoError(t, err)

			child := reqs.Clone()
			require.Equal(t, reqs.Doer, child.Doer)
			require.Equal(t, reqs.Method, child.Method)
			require.Equal(t, reqs.URL, child.URL)
			assert.EqualValues(t, reqs.Header, child.Header)

			if reqs.Header != nil {
				assert.EqualValues(t, reqs.Header, child.Header)
				reqs.Header.Add("K", "V")
				assert.Empty(t, child.Header.Get("K"), "child.header was a reference to original map, should be copy")
			} else {
				assert.Nil(t, child.Header)
			}

			assert.EqualValues(t, reqs.QueryParams, child.QueryParams)

			if len(reqs.QueryParams) > 0 {
				child.QueryParams.Set("color", "red")
				assert.Empty(t, reqs.QueryParams.Get("color"), "child.QueryParams should be a copy")
			}

			assert.Equal(t, reqs.Body, child.Body)
		})
	}
}

func TestRequester_Request_URLAndMethod(t *testing.T) {
	cases := []struct {
		options        []Option
		expectedMethod string
		expectedURL    string
	}{
		{[]Option{URL("http://a.io")}, "GET", "http://a.io"},
		{[]Option{RelativeURL("http://a.io")}, "GET", "http://a.io"},
		{[]Option{Get("http://a.io")}, "GET", "http://a.io"},
		{[]Option{Put("http://a.io")}, "PUT", "http://a.io"},
		{[]Option{URL("http://a.io/"), RelativeURL("foo")}, "GET", "http://a.io/foo"},
		{[]Option{URL("http://a.io/"), Post("foo")}, "POST", "http://a.io/foo"},
		// if relative relPath is an absolute url, base is ignored
		{[]Option{URL("http://a.io"), RelativeURL("http://b.io")}, "GET", "http://b.io"},
		{[]Option{RelativeURL("http://a.io"), RelativeURL("http://b.io")}, "GET", "http://b.io"},
		// last method setter takes priority
		{[]Option{Get("http://b.io"), Post("http://a.io")}, "POST", "http://a.io"},
		{[]Option{Post("http://a.io/"), Put("foo/"), Delete("bar")}, "DELETE", "http://a.io/foo/bar"},
		// last Base setter takes priority
		{[]Option{URL("http://a.io"), URL("http://b.io")}, "GET", "http://b.io"},
		// URL setters are additive
		{[]Option{URL("http://a.io/"), RelativeURL("foo/"), RelativeURL("bar")}, "GET", "http://a.io/foo/bar"},
		{[]Option{RelativeURL("http://a.io/"), RelativeURL("foo/"), RelativeURL("bar")}, "GET", "http://a.io/foo/bar"},
		// removes extra '/' between base and ref url
		{[]Option{URL("http://a.io/"), Get("/foo")}, "GET", "http://a.io/foo"},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			reqs, err := New(c.options...)
			require.NoError(t, err)
			req, err := reqs.RequestContext(context.Background())
			require.NoError(t, err)
			assert.Equal(t, c.expectedURL, req.URL.String())
			assert.Equal(t, c.expectedMethod, req.Method)
		})
	}

	t.Run("invalidmethod", func(t *testing.T) {
		b, err := New(Method("@"))
		require.NoError(t, err)
		req, err := b.RequestContext(context.Background())
		require.Error(t, err)
		require.Nil(t, req)
	})
}

func TestRequester_Request_QueryParams(t *testing.T) {
	cases := []struct {
		options     []Option
		expectedURL string
	}{
		{[]Option{URL("http://a.io"), QueryParams(paramsA)}, "http://a.io?limit=30"},
		{[]Option{URL("http://a.io/?color=red"), QueryParams(paramsA)}, "http://a.io/?color=red&limit=30"},
		{[]Option{URL("http://a.io"), QueryParams(paramsA), QueryParams(paramsB)}, "http://a.io?count=25&kind_name=recent&limit=30"},
		{[]Option{URL("http://a.io/"), RelativeURL("foo?relPath=yes"), QueryParams(paramsA)}, "http://a.io/foo?limit=30&relPath=yes"},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			reqs, err := New(c.options...)
			require.NoError(t, err)

			req, _ := reqs.RequestContext(context.Background())
			require.Equal(t, c.expectedURL, req.URL.String())
		})
	}
}

func TestRequester_Request_Body(t *testing.T) {
	cases := []struct {
		options             []Option
		expectedBody        string // expected Body io.Reader as a string
		expectedContentType string
	}{
		// Body (json)
		{[]Option{Body(modelA)}, `{"text":"note","favorite_count":12}`, ContentTypeJSON + ";charset=utf-8"},
		{[]Option{Body(&modelA)}, `{"text":"note","favorite_count":12}`, ContentTypeJSON + ";charset=utf-8"},
		{[]Option{Body(&FakeModel{})}, `{}`, ContentTypeJSON + ";charset=utf-8"},
		{[]Option{Body(FakeModel{})}, `{}`, ContentTypeJSON + ";charset=utf-8"},
		// BodyForm
		{[]Option{Form(), Body(paramsA)}, "limit=30", ContentTypeForm},
		{[]Option{Form(), Body(paramsB)}, "count=25&kind_name=recent", ContentTypeForm},
		{[]Option{Form(), Body(&paramsB)}, "count=25&kind_name=recent", ContentTypeForm},
		// Raw bodies, skips marshaler
		{[]Option{Body(strings.NewReader("this-is-a-test"))}, "this-is-a-test", ""},
		{[]Option{Body("this-is-a-test")}, "this-is-a-test", ""},
		{[]Option{Body([]byte("this-is-a-test"))}, "this-is-a-test", ""},
		// no body
		{nil, "", ""},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			reqs, err := New(c.options...)
			require.NoError(t, err)
			req, err := reqs.RequestContext(context.Background())
			require.NoError(t, err)

			if reqs.Body != nil {
				buf := new(bytes.Buffer)

				if _, err := buf.ReadFrom(req.Body); err != nil {
					t.Errorf("failed to read body: %v", err)
				}

				assert.Equal(t, c.expectedBody, buf.String())
			} else {
				assert.Nil(t, req.Body)
			}

			assert.Equal(t, c.expectedContentType, req.Header.Get(HeaderContentType))
		})
	}
}

func TestRequester_Request_Marshaler(t *testing.T) {
	var capturedV interface{}

	requester := Requester{
		Body: []string{"blue"},
		Marshaler: MarshalFunc(func(v interface{}) ([]byte, string, error) {
			capturedV = v
			return []byte("red"), "orange", nil
		}),
	}

	req, err := requester.RequestContext(context.Background())
	require.NoError(t, err)

	require.Equal(t, []string{"blue"}, capturedV)

	by, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	require.Equal(t, "red", string(by))
	require.Equal(t, "orange", req.Header.Get("Content-Type"))

	t.Run("errors", func(t *testing.T) {
		requester.Marshaler = MarshalFunc(func(v interface{}) ([]byte, string, error) {
			return nil, "", errors.New("boom") // nolint: err113
		})

		_, err := requester.RequestContext(context.Background())
		require.Error(t, err, "boom")
	})
}

func TestRequester_Request_ContentLength(t *testing.T) {
	reqs, err := New(Body("1234"))
	require.NoError(t, err)

	req, err := reqs.RequestContext(context.Background())
	require.NoError(t, err)

	// content length should be set automatically
	require.EqualValues(t, 4, req.ContentLength)

	// I should be able to override it
	reqs.ContentLength = 10

	req, err = reqs.RequestContext(context.Background())
	require.NoError(t, err)

	require.EqualValues(t, 10, req.ContentLength)
}

func TestRequester_Request_GetBody(t *testing.T) {
	reqs, err := New(Body("1234"))
	require.NoError(t, err)

	req, err := reqs.RequestContext(context.Background())
	require.NoError(t, err)

	// GetBody should be populated automatically
	rdr, err := req.GetBody()
	require.NoError(t, err)

	bts, err := io.ReadAll(rdr)
	require.NoError(t, err)

	require.Equal(t, "1234", string(bts))

	// I should be able to override it
	reqs.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("5678")), nil
	}

	req, err = reqs.RequestContext(context.Background())
	require.NoError(t, err)

	rdr, err = req.GetBody()
	require.NoError(t, err)

	bts, err = io.ReadAll(rdr)
	require.NoError(t, err)

	require.Equal(t, "5678", string(bts))
}

func TestRequester_Request_Host(t *testing.T) {
	reqs, err := New(URL("http://test.com/red"))
	require.NoError(t, err)

	req, err := reqs.RequestContext(context.Background())
	require.NoError(t, err)

	// Host should be set automatically
	require.Equal(t, "test.com", req.Host)

	// but I can override it
	reqs.Host = "test2.com"

	req, err = reqs.RequestContext(context.Background())
	require.NoError(t, err)

	require.Equal(t, "test2.com", req.Host)
}

func TestRequester_Request_TransferEncoding(t *testing.T) {
	reqs := Requester{}

	req, err := reqs.RequestContext(context.Background())
	require.NoError(t, err)

	// should be empty by default
	require.Nil(t, req.TransferEncoding)

	// but I can set it
	reqs.TransferEncoding = []string{"red"}

	req, err = reqs.RequestContext(context.Background())
	require.NoError(t, err)

	require.Equal(t, reqs.TransferEncoding, req.TransferEncoding)
}

func TestRequester_Request_Close(t *testing.T) {
	reqs := Requester{}

	req, err := reqs.RequestContext(context.Background())
	require.NoError(t, err)

	// should be false by default
	require.False(t, req.Close)

	// but I can set it
	reqs.Close = true

	req, err = reqs.RequestContext(context.Background())
	require.NoError(t, err)

	require.True(t, req.Close)
}

func TestRequester_Request_Trailer(t *testing.T) {
	reqs := Requester{}

	req, err := reqs.RequestContext(context.Background())
	require.NoError(t, err)

	// should be empty by default
	require.Nil(t, req.Trailer)

	// but I can set it
	reqs.Trailer = http.Header{"color": []string{"red"}}

	req, err = reqs.RequestContext(context.Background())
	require.NoError(t, err)

	require.Equal(t, reqs.Trailer, req.Trailer)
}

func TestRequester_Request_Header(t *testing.T) {
	reqs := Requester{}

	req, err := reqs.RequestContext(context.Background())
	require.NoError(t, err)

	// should be empty by default
	require.Empty(t, req.Header)

	// but I can set it
	reqs.Header = http.Header{"color": []string{"red"}}

	req, err = reqs.RequestContext(context.Background())
	require.NoError(t, err)

	require.Equal(t, reqs.Header, req.Header)
}

func TestRequester_Request_Context(t *testing.T) {
	reqs := Requester{}

	req, err := reqs.RequestContext(context.WithValue(context.Background(), colorContextKey, "red"))
	require.NoError(t, err)

	require.Equal(t, "red", req.Context().Value(colorContextKey))
}

func TestRequester_Request(t *testing.T) {
	reqs := Requester{}

	req, err := reqs.Request()
	require.NoError(t, err)

	require.NotNil(t, req)
}

func TestRequester_Request_options(t *testing.T) {
	reqs := Requester{}

	req, err := reqs.Request(Get("http://test.com/blue"))
	require.NoError(t, err)

	assert.Equal(t, "http://test.com/blue", req.URL.String())
}

func TestRequester_SendContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer ts.Close()

	i := Inspector{}
	r := MustNew(Get(ts.URL), &i)

	resp, err := r.SendContext(
		context.WithValue(context.Background(), colorContextKey, "purple"),
		Post("/server"),
	)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
	assert.Equal(t, "purple", i.Request.Context().Value(colorContextKey), "context should be passed through")
	assert.Equal(t, "GET", r.Method, "option arguments should have only affected that request, should not have leaked back into the Requester objects")

	t.Run("Send", func(t *testing.T) {
		resp, err := r.Send(Get("/server"))
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)
	})
}

func TestRequester_Receive_withopts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("blue")) // nolint: errcheck
	}))
	defer ts.Close()

	var called bool

	resp, _, err := MustNew(
		Get(ts.URL, "/profile"),
		UnmarshalFunc(func(data []byte, contentType string, v interface{}) error {
			called = true
			return nil
		}),
	).Receive(&testModel{})
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.True(t, called)
}

func TestRequester_ReceiveContext(t *testing.T) {
	mux := http.NewServeMux()

	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/model.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(206)
		w.Write([]byte(`{"color":"green","count":25}`)) // nolint: errcheck
	})

	mux.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"color":"red","count":30}`)) // nolint: errcheck
	})

	t.Run("success", func(t *testing.T) {
		cases := []struct {
			into interface{}
		}{
			{&testModel{}},
			{nil},
		}
		for _, c := range cases {
			t.Run(fmt.Sprintf("into=%v", c.into), func(t *testing.T) {
				i := Inspector{}

				resp, body, err := ReceiveContext(
					context.WithValue(context.Background(), colorContextKey, "purple"),
					c.into,
					Get(ts.URL, "/model.json"),
					&i,
				)
				require.NoError(t, err)

				defer resp.Body.Close()

				assert.Equal(t, 206, resp.StatusCode)
				assert.Equal(t, `{"color":"green","count":25}`, string(body))
				assert.Equal(t, "purple", i.Request.Context().Value(colorContextKey), "context should be passed through")

				if c.into != nil {
					assert.Equal(t, &testModel{"green", 25}, c.into)
				}
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		r := MustNew(
			Get(ts.URL, "/err"),
		)

		urlBefore := r.URL.String()
		resp, body, err := r.ReceiveContext(
			context.Background(),
			Get("/err"),
		)
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, `{"color":"red","count":30}`, string(body))
		assert.Equal(t, urlBefore, r.URL.String(), "the Get option should only affect the single request, it should not leak back into the Requester object")
	})

	// convenience functions which just wrap ReceiveContext
	t.Run("Receive", func(t *testing.T) {
		var m testModel

		resp, body, err := MustNew(Get(ts.URL, "/model.json")).Receive(&m)
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 206, resp.StatusCode)
		assert.Equal(t, `{"color":"green","count":25}`, string(body))
		assert.Equal(t, "green", m.Color)
	})

	t.Run("acceptoptionsforintoargs", func(t *testing.T) {
		mux.HandleFunc("/blue", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(208)
		})

		r := MustNew(Get(ts.URL, "/model.json"))

		// Receive will Options to be passed as the "into" arguments
		resp, _, err := r.Receive(Get("/blue"))
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 208, resp.StatusCode)

		// Options should be applied in the order of the arguments
		resp, _, err = r.Receive(Get("/red"), Get("/blue"))
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 208, resp.StatusCode)

		// variants
		ctx := context.Background()
		resp, _, err = r.ReceiveContext(ctx, Get("/blue"))
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 208, resp.StatusCode)
	})
}

func TestRequester_Params(t *testing.T) {
	reqr := &Requester{}
	reqr.Params().Set("color", "red")
	assert.Equal(t, "red", reqr.QueryParams.Get("color"))
}

func TestRequester_Headers(t *testing.T) {
	reqr := &Requester{}
	reqr.Headers().Set("color", "red")
	assert.Equal(t, "red", reqr.Header.Get("color"))
}

func TestRequester_Trailers(t *testing.T) {
	reqr := &Requester{}
	reqr.Trailers().Set("color", "red")
	assert.Equal(t, "red", reqr.Trailer.Get("color"))
}

type TestStruct struct {
	Color     string
	Count     int
	Flavor    string
	Important bool
}

func BenchmarkRequester_Receive(b *testing.B) {
	inputJSON := `{"color":"blue","count":10,"flavor":"vanilla","important":true}`

	h := map[string][]string{"Content-Type": {"application/json"}, "Content-Length": {strconv.Itoa(len([]byte(inputJSON)))}}

	var mockServer DoerFunc = func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: 200,
			Header:     h,
			Body:       io.NopCloser(strings.NewReader(inputJSON)),
		}

		return resp, nil
	}

	// smoke test
	var ts TestStruct

	resp, s, err := Receive(&ts, mockServer, JSON(false), Get("/test"))
	require.NoError(b, err)

	defer resp.Body.Close()

	require.JSONEq(b, inputJSON, string(s))
	require.Equal(b, TestStruct{Color: "blue", Count: 10, Flavor: "vanilla", Important: true}, ts)

	b.Run("simple", func(b *testing.B) {
		b.Run("requester", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resp, _, err := Receive(&TestStruct{}, mockServer, Get("/test"))
				require.NoError(b, err)

				resp.Body.Close()
			}
		})

		b.Run("base", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest("GET", "/test", nil)

				resp, err := mockServer.Do(req)
				require.NoError(b, err)

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				if err := json.Unmarshal(body, &TestStruct{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("complex", func(b *testing.B) {
		b.Run("requester", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resp, _, err := Receive(&ts,
					mockServer,
					Get("/test/blue/green"),
					JSON(false),
					Header("X-Under", "Over"),
					Header("X-Over", "Under"),
					QueryParam("color", "blue"),
					QueryParam("q", "user=sam"),
					Body(&ts),
				)
				require.NoError(b, err)

				resp.Body.Close()
			}
		})

		b.Run("requester_attrs", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r := MustNew(
					mockServer,
					Get("/test/blue/green"),
					JSON(false),
					QueryParam("color", "blue"),
					QueryParam("q", "user=sam"),
					Body(&ts),
				)
				r.Header.Add("X-Under", "Over")
				r.Header.Add("X-Over", "Under")

				resp, _, err := r.Receive(&ts)
				require.NoError(b, err)

				resp.Body.Close()
			}
		})

		b.Run("base", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reqbody, _ := json.Marshal(&ts)

				qp := url.Values{}
				qp.Add("color", "blue")
				qp.Add("q", "user=sam")

				req, _ := http.NewRequestWithContext(context.Background(), "GET", "/test/blue/green?"+qp.Encode(), bytes.NewReader(reqbody))
				req.Header.Set("X-Under", "Over")
				req.Header.Set("X-Over", "Under")
				req.Header.Set("Content-Type", "application/json")

				resp, err := mockServer.Do(req)
				require.NoError(b, err)

				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				require.NoError(b, err)

				err = json.Unmarshal(body, &TestStruct{})
				require.NoError(b, err)
			}
		})
	})
}

func ExampleRequester_Receive() {
	r := MustNew(MockDoer(200,
		Body("red"),
	))

	resp, body, _ := r.Receive(Get("http://api.com/resource"))

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode, string(body))
}

func ExampleRequester_Receive_unmarshal() {
	type Resource struct {
		Color string `json:"color"`
	}

	r := MustNew(MockDoer(200,
		JSON(true),
		Body(Resource{Color: "red"}),
	))

	var resource Resource

	resp, body, _ := r.Receive(&resource, Get("http://api.com/resource"))

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
	fmt.Println(resource.Color)
}

func ExampleRequester_Request() {
	r := MustNew(
		Get("http://api.com/resource"),
		Header("X-Color", "red"),
		QueryParam("flavor", "vanilla"),
	)

	req, _ := r.Request(
		JSON(true),
		Body(map[string]interface{}{"size": "big"}),
	)

	fmt.Printf("%s %s %s\n", req.Method, req.URL.String(), req.Proto)
	fmt.Println(HeaderContentType+":", req.Header.Get(HeaderContentType))
	fmt.Println(HeaderAccept+":", req.Header.Get(HeaderAccept))
	fmt.Println("X-Color:", req.Header.Get("X-Color"))

	if _, err := io.Copy(os.Stdout, req.Body); err != nil {
		fmt.Println(err)
	}
}

func ExampleRequester_Send() {
	r := MustNew(MockDoer(204))

	resp, _ := r.Send(Get("resources/1"))

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
}
