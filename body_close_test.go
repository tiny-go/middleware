package mw

import (
	"net/http"
	"testing"
)

type mockReadCloser struct{ closed bool }

func (mrc *mockReadCloser) Read([]byte) (i int, e error) { return }

func (mrc *mockReadCloser) Close() (e error) { mrc.closed = true; return }

func Test_BodyClose(t *testing.T) {
	t.Run("body close middleware should close request body after each request", func(t *testing.T) {
		handler := BodyClose(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		body := &mockReadCloser{}
		if body.closed {
			t.Errorf("body should not be closed before calling the hadler")
		}
		handler.ServeHTTP(nil, &http.Request{Body: body})
		if !body.closed {
			t.Errorf("body should be closed after calling the handler")
		}
	})
}
