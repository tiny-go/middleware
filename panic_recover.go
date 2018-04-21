package mw

import (
	"fmt"
	"net/http"
)

// PanicHandler reports the error (panic) to the client as an HTTP error. If HTTP
// error status code can be retrieved PanicHandler sends provided code to the client
// instead of default (500). If provided panic (value) does not implement any of
// supported interfaces (error/Error) - it tries to convert a panic to a string.
func PanicHandler(w http.ResponseWriter, p interface{}) {
	switch e := p.(type) {
	case nil:
		// ignore (panics that throw nil can be used to indicate that handler finished
		// the task successfully and all the next handlers can be igored)
	case Error:
		// retrieve status code and error message
		http.Error(w, e.Error(), e.Code())
	case error:
		// all standard errors without codes will be sent as Internal Server Error
		http.Error(w, e.Error(), http.StatusInternalServerError)
	default:
		// everything else
		http.Error(w, fmt.Sprint(p), http.StatusInternalServerError)
	}
}

// PanicRecover returns a middleware that recovers from the panic.
func PanicRecover(onFail func(http.ResponseWriter, interface{})) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// recover from the panic and report
				defer func() {
					if r := recover(); r != nil {
						onFail(w, r)
					}
				}()
				// call next middleware
				next.ServeHTTP(w, r)
			},
		)
	}
}
