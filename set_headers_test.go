package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_SetHeaders(t *testing.T) {
	t.Run("set headers middleware should set provided response headers", func(t *testing.T) {
		headers := map[string]string{
			"Content-Encoding": "gzip",
			"Content-Length":   "27225",
			"Content-Type":     "application/json",
		}
		handler := SetHeaders(headers)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, nil)

		for key, value := range headers {
			if w.Header().Get(key) != value {
				t.Errorf("header %q expected to be %q but got %q", key, value, w.Header().Get(key))
			}
		}
	})
}
