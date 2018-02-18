# go-middleware

[![GoDoc][godoc-badge]][godoc-link]
[![Build Status][circleci-badge]][circleci-link]
[![Report Card][report-badge]][report-link]
[![GoCover][cover-badge]][cover-link]

Golang HTTP middleware

### Installation
Since `go-middleware` is an open source project and repository is public you can simply install it with `go get`:
```bash
$ go get github.com/Alma-media/go-middleware
```

### Currently available middleware
- `BodyClose` - closes request body for each request
- `ContextDeadline` - sets request timeout (demands additional logic in your app)
- `PanicRecover` - catches the panics inside our chain
- `SetHeaders` - provides an easy way to set response headers

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

    	"github.com/Alma-media/go-middleware"
    )

    var (
        panicHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
    	    panic("something went wrong")
        }
    )

    func main() {
    	http.Handle(
    		"/",
    		mw.
    			New(mw.PanicRecover(log.New(os.Stdout, "", 0))).
    			Use(mw.BodyClose).
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
    		mw.Chain(
    			mw.PanicRecover(log.New(os.Stdout, "", 0)),
    			mw.BodyClose,
    			panicHandler,
    		),
    	)
    	log.Fatal(http.ListenAndServe(":8080", nil))
    }
    ```

[godoc-badge]: https://godoc.org/github.com/Alma-media/go-middleware?status.svg
[godoc-link]: https://godoc.org/github.com/Alma-media/go-middleware
[circleci-badge]: https://circleci.com/gh/Alma-media/go-middleware.svg?style=shield
[circleci-link]: https://circleci.com/gh/Alma-media/go-middleware
[report-badge]: https://goreportcard.com/badge/github.com/Alma-media/go-middleware
[report-link]: https://goreportcard.com/report/github.com/Alma-media/go-middleware
[cover-badge]: https://gocover.io/_badge/github.com/Alma-media/go-middleware
[cover-link]: https://gocover.io/github.com/Alma-media/go-middleware
