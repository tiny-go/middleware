package mw

import (
	"context"
	"net/http"
	"time"
)

// ContextDeadline adds timeout to request's context.
func ContextDeadline(timeout time.Duration) Middleware {
	// create a new Middleware
	return func(next http.Handler) http.Handler {
		// define the httprouter.Handle
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ctx is the Context for this handler, calling cancel closes the ctx.Done
			// channel, which is the cancellation signal for requests started by this handler
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)
			if timeout > 0 {
				// the request has a timeout, so create a context that is canceled automatically
				// when the timeout expires
				ctx, cancel = context.WithTimeout(context.Background(), timeout)
			} else {
				ctx, cancel = context.WithCancel(context.Background())
			}
			defer cancel()
			// replace request context
			r = r.WithContext(ctx)
			// call next handler
			next.ServeHTTP(w, r)
		})
	}
}
