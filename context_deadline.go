package mw

import (
	"context"
	"net/http"
	"time"
)

// Go function runs the handler and processes its execuion in a new goroutine.
// It also passes Done channel to the handler. Once channel is closed handler
// should be able to stop its goroutine to avoid resource leaks.
func Go(ctx context.Context, handler func(stop <-chan struct{}) error) error {
	// error chan
	errChan := make(chan error, 1)
	// call handler in goroutine
	go func() { errChan <- handler(ctx.Done()) }()
	// wait until context deadline or job is done
	select {
	// job was done
	case err := <-errChan:
		return err
	// timeout
	case <-ctx.Done():
		// send context deadline
		return ctx.Err()
	}
}

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
				ctx, cancel = context.WithTimeout(r.Context(), timeout)
			} else {
				ctx, cancel = context.WithCancel(r.Context())
			}
			defer cancel()
			// replace request context
			r = r.WithContext(ctx)
			// call next handler
			next.ServeHTTP(w, r)
		})
	}
}
