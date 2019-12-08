package mw

import (
	"context"
	"net/http"
)

// StatusNoResponse is returned when request is canceled
const StatusNoResponse = 444

// ContextHandler reads from context.Done channel to handle deadline/timeout
func ContextHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		processed := make(chan struct{})
		go func() {
			defer close(processed)
			next.ServeHTTP(w, r)
		}()
		select {
		case <-r.Context().Done():
			switch r.Context().Err() {
			case nil:
				// do nothing
			case context.Canceled:
				http.Error(w, r.Context().Err().Error(), StatusNoResponse)
			case context.DeadlineExceeded:
				http.Error(w, r.Context().Err().Error(), http.StatusRequestTimeout)
			default:
				// handle unknown errors
				http.Error(w, r.Context().Err().Error(), http.StatusInternalServerError)
			}
			return
		case <-processed:
			return
		}
	})
}
