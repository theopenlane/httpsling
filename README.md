[![Build status](https://badge.buildkite.com/f74a461120ffcadbf7796d5aac8ae8c03a1cbcfda142220074.svg)](https://buildkite.com/theopenlane/httpsling?branch=main)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=theopenlane_httpsling&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=theopenlane_httpsling)
[![Go Report Card](https://goreportcard.com/badge/github.com/theopenlane/httpsling)](https://goreportcard.com/report/github.com/theopenlane/httpsling)
[![Go Reference](https://pkg.go.dev/badge/github.com/theopenlane/httpsling.svg)](https://pkg.go.dev/github.com/theopenlane/httpsling)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)


# Slinging HTTP

The `httpsling` library simplifies the way you make HTTP requests. It's intended to provide an easy-to-use interface for sending requests and handling responses, reducing the boilerplate code typically associated with the `net/http` package.

## Overview

Creating a new `Requester` and making a request should be straightforward:

```go
package main

import (
	"log"
	"net/http"

	"github.com/theopenlane/httpsling"
)

func main() {
	requester, err := httpsling.New(
		httpsling.Client(), // use the default sling client
		httpsling.URL("https://api.example.com"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Perform a GET request
	var out map[string]any
	resp, err := requester.Receive(&out, httpsling.Get("resource"))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	log.Println(out)
}
```

### Core Functions

```go
// build just the request
Request(...Option) (*http.Request, error)
RequestWithContext(context.Context, ...Option) (*http.Request, error)

// build the request and send the request
Send(...Option) (*http.Response, error)
SendWithContext(context.Context, ...Option) (*http.Response, error)

// build and send the request and parse the response into a value
Receive(any, ...Option) (*http.Response, error)
ReceiveWithContext(context.Context, any, ...Option) (*http.Response, error)

// typed helper using generics
ReceiveInto[T any](...Option) (*http.Response, T, error)
ReceiveIntoWithContext[T any](context.Context, ...Option) (*http.Response, T, error)

// stream response body into an io.Writer
ReceiveTo(io.Writer, ...Option) (*http.Response, int64, error)
ReceiveToWithContext(context.Context, io.Writer, ...Option) (*http.Response, int64, error)
```

### Configuring BaseURL

Set the base URL for all requests using the `Requester` option `URL`:

```go
 httpsling.URL("https://api.example.com"),
```

### Setting Headers

Set default headers for all requests, e.g. Bearer token Authorization

```go
    requester.Apply(httpsling.BearerAuth("YOUR_ACCESS_TOKEN"))
```

### Setting up CookieJar

Add a Cookie Jar to the default client:

```go
    requester, err = httpsling.New(
        httpsling.Client(
            httpclient.CookieJar(nil), // Use a cookie jar to store cookies
        ),
    )
    if err != nil {
        return nil, err
    }
```

### Configuring Timeouts

Define a global timeout for all requests to prevent indefinitely hanging operations:

```go
    httpsling.Client(
        httpclient.Timeout(time.Duration(30*time.Second)),
    ),
```

### TLS Configuration

Custom TLS configurations can be applied for enhanced security measures, such as loading custom certificates:

```go
    httpsling.Client(
        httpclient.SkipVerify(true),
    ),
```

## Requests

The library provides a `Receive` to construct and dispatch HTTP. Here are examples of performing various types of requests, including adding query parameters, setting headers, and attaching a body to your requests.

#### GET Request

```go
    resp, err := requester.ReceiveWithContext(context.Background(), &out,
        httpsling.Get("/path"),
        httpsling.QueryParam("query", "meow"),
    )
```

#### POST Request

```go
    resp, err := requester.ReceiveWithContext(context.Background(), &out,
        httpsling.Post("/path"),
        httpsling.Body(map[string]any{"key": "value"})
    )
```

#### PUT Request

```go
    resp, err := requester.ReceiveWithContext(context.Background(), &out,
        httpsling.Put("/path/123456"),
        httpsling.Body(map[string]any{"key": "newValue"})
    )
```

#### DELETE Request

```go
    resp, err := requester.ReceiveWithContext(context.Background(), &out,
        httpsling.Delete("/path/123456"),
    )
```

### Authentication

Supports various authentication methods:

- **Basic Auth**:

```go
    requester.Apply(httpsling.BasicAuth("username", "superSecurePassword!"))
```

- **Bearer Token**:

```go
    requester.Apply(httpsling.BearerAuth("YOUR_ACCESS_TOKEN"))
```

## Responses

Handling responses is necessary in determining the outcome of your HTTP requests - the library has some built-in response code validators and other tasty things.

```go
type APIResponse struct {
    Data string `json:"data"`
}

var out APIResponse
resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post("/path"),
		httpsling.Body(in))

defer resp.Body.Close()

log.Printf("Status Code: %d\n", resp.StatusCode)
log.Printf("Response Data: %s\n", out.Data)
```

Note: Receive fully reads the response into memory to unmarshal, and restores `resp.Body` so it remains readable by callers.

### Evaluating Response Success

To assess whether the HTTP request was successful:

- **IsSuccess**: Check if the status code signifies a successful response

```go
    if httpsling.IsSuccess(resp) {
        fmt.Println("The request succeeded hot diggity dog")
    }
```

## Inspirations

This library was inspired by and built upon the work of several other HTTP client libraries:

- [Dghubble/sling](https://github.com/dghubble/sling)
- [Monaco-io/request](https://github.com/monaco-io/request)
- [Go-resty/resty](https://github.com/go-resty/resty)
- [Fiber Client](https://github.com/gofiber/fiber)

## Contributing

See [contributing](.github/CONTRIBUTING.md) for details.
