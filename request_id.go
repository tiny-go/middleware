package mw

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
)

const requestIDHeader = "X-Request-Id"

// requestIDKey is a private unique key that is used for request ID in the context.
type requestIDKey struct{ kind string }

// RequestID is a middleware that injects a request ID into the context of each
// request. A request ID is a string (randomly generated UUID).
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestUUID, _ := uuid.NewV4()
				requestID = requestUUID.String()
			}
			next.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), requestIDKey{}, requestID)),
			)
		},
	)
}

// RequestIDFromContext pulls request ID from the context or empty string.
func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}
