package httptestutil_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httptestutil"
)

func Example() {
	mux := http.NewServeMux()
	mux.Handle("/echo", httpsling.MockHandler(201, httpsling.Body("pong")))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// inspect server traffic
	is := httptestutil.Inspect(ts)

	// construct a pre-configured Requester
	r := httptestutil.Requester(ts)

	var out string
	resp, _ := r.Receive(&out, httpsling.Get("/echo"), httpsling.Body("ping"))

	ex := is.LastExchange()
	fmt.Println("server received: " + ex.RequestBody.String())
	fmt.Println("server sent: " + strconv.Itoa(ex.StatusCode))
	fmt.Println("server sent: " + ex.ResponseBody.String())
	fmt.Println("client received: " + strconv.Itoa(resp.StatusCode))
	fmt.Println("client received: " + fmt.Sprintf("%s", out))

	// Output:
	// server received: ping
	// server sent: 201
	// server sent: pong
	// client received: 201
	// client received: pong
}
