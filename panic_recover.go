package middleware

import (
	"net/http"
)

// OnPanic is a function that should contain the logic responsible for any kind
// of error/panic reporting (to the console, file, directly to the client or simply
// ignore them).
type OnPanic func(http.ResponseWriter, interface{})

// PanicRecover returns a middleware that recovers from the panic.
//
// A trivial example (retrieving an error from the panic and sending to the client
// using errors package) is:
//
//  package main
//
//  import (
//      "log"
//      "net/http"
//
//      "github.com/tiny-go/errors"
//      "github.com/tiny-go/middleware"
//  )
//
//  var (
//      panicHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
//          panic("something went wrong")
//      }
//  )
//
//  func main() {
//      http.Handle(
//        "/",
//        mw.
//            // with HTTP panic handler
//            New(mw.PanicRecover(errors.Send)).
//            Then(panicHandler),
//      )
//      log.Fatal(http.ListenAndServe(":8080", nil))
//  }
func PanicRecover(onPanic OnPanic) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// recover from panic and call the panic handler
				defer func() { onPanic(w, recover()) }()
				// call next middleware
				next.ServeHTTP(w, r)
			},
		)
	}
}
