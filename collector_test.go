package mw

import (
	"net/http"
	"testing"
)

var (
	mwOne Middleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	mwTwo Middleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	mwThree Middleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
)

func Test_List(t *testing.T) {
	t.Run("Given a middleware List", func(t *testing.T) {
		list := NewList()
		if len(list.middleware) != 0 {
			t.Error("middleware map should be empty after initialization")
		}
		t.Run("register middleware for HTTP method", func(t *testing.T) {
			list.AddMiddleware(http.MethodGet, mwOne)
			mw, ok := list.middleware[http.MethodGet]
			if !ok {
				t.Errorf("middleware map is expected to contain key %q", http.MethodGet)
			}
			if len(mw) != 1 {
				t.Errorf("middleware list for key %q should contain exactly one middleware", http.MethodGet)
			}
		})
		t.Run("add middleware func(s) to existing chain", func(t *testing.T) {
			list.AddMiddleware(http.MethodGet, mwTwo, mwThree)
			mw, ok := list.middleware[http.MethodGet]
			if !ok {
				t.Errorf("middleware map is expected to contain key %q", http.MethodGet)
			}
			if len(mw) != 3 {
				t.Errorf("middleware list for key %q should contain three middleware funcs", http.MethodGet)
			}
		})
		t.Run("get an empty middleware chain (by default)", func(t *testing.T) {
			mw := list.Middleware(http.MethodPost)
			if len(mw) != 0 {
				t.Errorf("middleware list should be empty for method %q", http.MethodPost)
			}
		})
		t.Run("get middleware registered for HTTP method", func(t *testing.T) {
			mw := list.Middleware(http.MethodGet)
			if mw[0] == nil {
				t.Error("mwOne should not be nil")
			}
			if mw[1] == nil {
				t.Error("mwTwo should not be nil")
			}
			if mw[2] == nil {
				t.Error("mwThree should not be nil")
			}
		})
	})
}
