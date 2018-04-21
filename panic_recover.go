package mw

import (
	"fmt"
	"net/http"
)

// PanicHandler reports the error (panic). Panic can be reported to the client as
// an HTTP error if handler is able to obtain status code or at least the error
// message, otherwice it tries to convert a panic to a string.
func PanicHandler(w http.ResponseWriter, p interface{}) {
	switch e := p.(type) {
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
