package mw

import (
	"net/http"
)

// BodyClose is a middleware that closes the request body after each request.
func BodyClose(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// close request body on exit
			defer r.Body.Close()
			// call next middleware func
			next.ServeHTTP(w, r)
		},
	)
}
