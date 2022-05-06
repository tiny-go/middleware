package middleware

import (
	"net/http"
	"time"
)

// Throttle provides a middleware that limits the amount of messages processed
// per unit of time. Note that middleware func can be defined globally to be shared
// by multiple handlers
//
// Example:
//
//  var mw = Throttle(10, time.Second)
//
//  http.Handle("/one", mw.Then(handlerOne))
//  http.Handle("/two", mw.Then(handlerTwo))
func Throttle(count int64, duration time.Duration) Middleware {
	ticker := time.Tick(duration / time.Duration(count))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-ticker
			next.ServeHTTP(w, r)
		})
	}
}
