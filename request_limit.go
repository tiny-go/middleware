package mw

import (
	"net/http"

	"github.com/tiny-go/errors"
)

// RequestLimiter middleware apply concurrent request limit
// TODO: provide a choise either "wait" or "send error"
func RequestLimiter(fn errors.HandlerFunc, maxConcurrentRequests int) Middleware {
	if fn == nil {
		fn = http.Error
	}
	limitChan := make(chan struct{}, maxConcurrentRequests)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case limitChan <- struct{}{}:
				defer func() { <-limitChan }()
				next.ServeHTTP(w, r)
			default:
				fn(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			}
		})
	}
}
