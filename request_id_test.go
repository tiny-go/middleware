package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
)

func Test_ReqiestID(t *testing.T) {
	t.Run("request ID middleware should extract request ID from headers or generate a new one", func(t *testing.T) {
		testCases := []struct {
			title   string
			request *http.Request
			handler http.Handler
		}{
			{
				title: "should return an empty request ID when middleware is not used",
				request: func() *http.Request {
					request, _ := http.NewRequest("", "", nil)
					return request
				}(),
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestID := RequestIDFromContext(r.Context())
					if requestID != "" {
						t.Error("request ID should not be set")
					}
				}),
			},
			{
				title: "should retrieve request ID from headers when provided by the client",
				request: func() *http.Request {
					request, _ := http.NewRequest("", "", nil)
					request.Header.Set(requestIDHeader, "ID-provided-by-client")
					return request
				}(),
				handler: RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestID := RequestIDFromContext(r.Context())
					if requestID != "ID-provided-by-client" {
						t.Errorf("request ID has unexpected value: %s", requestID)
					}
				})),
			},
			{
				title: "should retrieve generate new request ID when not provided by the client",
				request: func() *http.Request {
					request, _ := http.NewRequest("", "", nil)
					return request
				}(),
				handler: RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestID := RequestIDFromContext(r.Context())
					if _, err := uuid.FromString(requestID); err != nil {
						t.Errorf("request ID should be valid UUID: %s", err)
					}
				})),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				tc.handler.ServeHTTP(httptest.NewRecorder(), tc.request)
			})
		}
	})
}
