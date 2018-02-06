package mw

import (
	"net/http"
)

// PanicRecover returns a middleware that recovers from the panic.
func PanicRecover(logger interface {
	Println(v ...interface{})
}) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// recover from the panic and report
				defer func() {
					if r := recover(); r != nil {
						logger.Println("Recovered from panic:", r)
					}
				}()
				// call next middleware
				next.ServeHTTP(w, r)
			},
		)
	}
}
