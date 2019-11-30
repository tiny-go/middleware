package mw

import "net/http"

// TODO: provide a choise either "wait" or "send error"

// ErrHandler is a function that handles errors, for instance http.Error
type ErrHandler func(w http.ResponseWriter, message string, code int)

// RequestLimiter middleware apply concurrent request limit
func RequestLimiter(fn ErrHandler, maxConcurrentRequests int) Middleware {
	limitChan := make(chan struct{}, maxConcurrentRequests)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case limitChan <- struct{}{}:
				defer func() { <-limitChan }()
				next.ServeHTTP(w, r)
			default:
				fn(w, "too many requests", http.StatusTooManyRequests)
			}
		})
	}
}
