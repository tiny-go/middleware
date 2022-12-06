# go-middleware

[![GoDoc][godoc-badge]][godoc-link]
[![License][license-badge]][license-link]
[![Report Card][report-badge]][report-link]
[![GoCover][cover-badge]][cover-link]

Golang HTTP middleware

### Installation
Since `go-middleware` is an open source project and repository is public you can simply install it with `go get`:
```bash
$ go get github.com/tiny-go/middleware
```

### Currently available middleware
- `BodyClose` - closes request body for each request
- `ContextDeadline` - sets request timeout (demands additional logic in your app)
- `PanicRecover` - catches the panics inside our chain, can be used as error handler (similar to `try/catch`) with corresponding panic handler
- `SetHeaders` - provides an easy way to set response headers
- `JwtHS256` - verifies JWT (JSON Web Token) signed with HMAC signing method and parses its body to the provided receiver that is going to be available to next handlers through the request context
- `Codec` - searches for suitable request/response codecs according to "Content-Type"/"Accept" headers and puts  them into the context

### Experimental middleware
It means that work is still in progress, a lot of things can be changed or even completely removed
- `AsyncRequest` - allows to set `request timeout` (for HTTP request) and `async timeout` (for background execution), if request has not been processed during `request timeout` - middleware returns `request ID` and HTTP code 202 (`Accepted`). You can make a new request wtih given `request ID` later to obtain the result. The result can be provided only once and won't be available after that anymore. If handler did not finish its task during `async timeout` - middleware sends an HTTP error with code 408 (`RequestTimeout`) executing next async request with current `request ID`.

### Examples

1. Build the handler using middleware chaining functions:
- `New()` - start the chain. Can accept 0 (zero) or more arguments.
- `Use()` - add middleware to existing chain.
- `Then()` - set the final handler (which is `http.Handler`).

    ```go
    package main

    import (
    	"log"
    	"net/http"
    	"os"

    	"github.com/tiny-go/middleware"
    	"github.com/tiny-go/codec/driver"
    	"github.com/tiny-go/codec/driver/json"
    	"github.com/tiny-go/codec/driver/xml"
    )

    var (
        panicHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
    	    panic("something went wrong")
        }
    )

    func main() {
    	http.Handle(
    		"/",
    		middleware.
    			// with HTTP panic handler
    			New(middleware.PanicRecover(middleware.PanicHandler)).
    			Use(middleware.BodyClose).
    			Use(middleware.Codec(driver.DummyRegistry{&json.JSON{}, &xml.XML{}})).
    			Then(panicHandler),
    	)
    	log.Fatal(http.ListenAndServe(":8080", nil))
    }

    ```

2. Build the handler with `Chain()` (variadic) func which accepts the next list of argument types:
- `http.Handler` and everything else that implements this interface (for instance, `http.HandlerFunc`)
- `func(w http.ResponseWriter, r *http.Request)`
- `MiddlewareFunc`/`func(http.ResponseWriter, *http.Request, http.Handler)`
- `Middleware`/`func(http.Handler) http.Handler`

    There is no sense to provide entire example since import and variable declaration sections are going to be the same, only `main()` func is going to be changed:

    ```go
    func main() {
    	http.Handle(
    		"/",
    		middleware.Chain(
    			// with custom panic handler
    			middleware.PanicRecover(func(_ http.ResponseWriter, r interface{}) {
    				if r != nil {
    					log.Println(r)
    				}
    			}),
    			middleware.BodyClose,
    			panicHandler,
    		),
    	)
    	log.Fatal(http.ListenAndServe(":8080", nil))
    }
    ```

[godoc-badge]: https://godoc.org/github.com/tiny-go/middleware?status.svg
[godoc-link]: https://godoc.org/github.com/tiny-go/middleware
[license-badge]: https://img.shields.io/:license-MIT-green.svg
[license-link]: https://opensource.org/licenses/MIT
[report-badge]: https://goreportcard.com/badge/github.com/tiny-go/middleware
[report-link]: https://goreportcard.com/report/github.com/tiny-go/middleware
[cover-badge]: https://gocover.io/_badge/github.com/tiny-go/middleware
[cover-link]: https://gocover.io/github.com/tiny-go/middleware
