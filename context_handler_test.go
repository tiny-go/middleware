package mw

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_ContextExceeder(t *testing.T) {
	t.Run("should send an error when the context is canceled", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("", "", nil)

		ctx, fn := context.WithCancel(context.Background())
		r = r.WithContext(ctx)
		fn()

		handler := ContextHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second)
		}))
		handler.ServeHTTP(w, r)
		if w.Code != StatusNoResponse {
			t.Errorf("response code was expected to be 444, got %d", w.Code)
		}
		if w.Body.String() != "context canceled\n" {
			t.Errorf("the response %q was expected to be \"context canceled\n\"", w.Body.String())
		}
	})

	t.Run("should send an error when the context deadline exceeded", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("", "", nil)

		ctx, fn := context.WithDeadline(context.Background(), time.Time{})
		r = r.WithContext(ctx)
		defer fn()

		handler := ContextHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second)
		}))
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusRequestTimeout {
			t.Errorf("response code was expected to be 408, got %d", w.Code)
		}
		if w.Body.String() != "context deadline exceeded\n" {
			t.Errorf("the response %q was expected to be \"context canceled\n\"", w.Body.String())
		}
	})

	t.Run("should not return an error if request was processed in given time", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("", "", nil)

		handler := ContextHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("success\n"))
		}))
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("response code was expected to be 200, got %d", w.Code)
		}
		if w.Body.String() != "success\n" {
			t.Errorf("the response %q was expected to be \"success\n\"", w.Body.String())
		}
	})
}
