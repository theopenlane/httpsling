package httpsling_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
	"github.com/theopenlane/httpsling/httptestutil"
)

func Example() {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"color":"red"}`))
	}))
	defer s.Close()

	var out map[string]string
	resp, _ := Receive(
		out,
		Get(s.URL),
	)

	fmt.Println(resp.StatusCode)
	fmt.Printf("%s", out)
}

func Example_receive() {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"color":"red"}`))
	}))
	defer s.Close()

	r := struct {
		Color string `json:"color"`
	}{}

	Receive(&r, Get(s.URL))

	fmt.Println(r.Color)
}

func Example_everything() {
	type Resource struct {
		ID    string `json:"id"`
		Color string `json:"color"`
	}

	s := httptest.NewServer(MockHandler(201,
		JSON(true),
		Body(&Resource{Color: "red", ID: "123"}),
	))
	defer s.Close()

	r := httptestutil.Requester(s,
		Post("/resources?size=big"),
		BearerAuth("atoken"),
		JSON(true),
		Body(&Resource{Color: "red"}),
		ExpectCode(201),
		Header("X-Request-Id", "5"),
		QueryParam("flavor", "vanilla"),
		QueryParams(&struct {
			Type string `url:"type"`
		}{Type: "upload"}),
		Client(
			httpclient.SkipVerify(true),
			httpclient.Timeout(5*time.Second),
			httpclient.MaxRedirects(3),
		),
	)

	r.MustApply(DumpToStderr())
	httptestutil.Dump(s, os.Stderr)

	serverInspector := httptestutil.Inspect(s)
	clientInspector := Inspect(r)

	var resource Resource

	resp, err := r.Receive(&resource)
	if err != nil {
		panic(err)
	}

	fmt.Println("client-side request url path:", clientInspector.Request.URL.Path)
	fmt.Println("client-side request query:", clientInspector.Request.URL.RawQuery)
	fmt.Println("client-side request body:", clientInspector.RequestBody.String())

	ex := serverInspector.LastExchange()
	fmt.Println("server-side request authorization header:", ex.Request.Header.Get("Authorization"))
	fmt.Println("server-side request request body:", ex.RequestBody.String())
	fmt.Println("server-side request response body:", ex.ResponseBody.String())

	fmt.Println("client-side response body:", clientInspector.ResponseBody.String())

	fmt.Println("response status code:", resp.StatusCode)
	fmt.Println("unmarshaled response body:", resource)
}
