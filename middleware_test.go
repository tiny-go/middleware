package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	middlewareOne = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("/mw1 before next"))
			next.ServeHTTP(w, r)
			w.Write([]byte("/mw1 after next"))
		})
	}

	middlewareTwo = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("/mw2 before next"))
			next.ServeHTTP(w, r)
			w.Write([]byte("/mw2 after next"))
		})
	}

	middlewareThree = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("/mw3 before next"))
			next.ServeHTTP(w, r)
			w.Write([]byte("/mw3 after next"))
		})
	}

	middlewareFuncOne = func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		w.Write([]byte("/mw func1 before next"))
		next.ServeHTTP(w, r)
		w.Write([]byte("/mw func1 after next"))
	}

	middlewareFuncTwo = func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		w.Write([]byte("/mw func2 before next"))
		next.ServeHTTP(w, r)
		w.Write([]byte("/mw func2 after next"))
	}

	middlewareBreak Middleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("/skip the rest"))
		})
	}

	handlerOne = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/first handler"))
	}

	handlerTwo = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/second handler"))
	}

	handlerFinal = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/final handler"))
	})
)

func Test_Middleware(t *testing.T) {
	type testCase struct {
		title   string
		handler http.Handler
		out     string
	}

	cases := []testCase{
		{
			title:   "build handler with single middleware (one call of Use() func with single argument)",
			handler: New().Use(middlewareOne).Then(handlerFinal),
			out:     "/mw1 before next/final handler/mw1 after next",
		},
		{
			title:   "build handler passing middleware to the constructor (call New() with arguments)",
			handler: New(middlewareOne, middlewareTwo).Use(middlewareThree).Then(handlerFinal),
			out:     "/mw1 before next/mw2 before next/mw3 before next/final handler/mw3 after next/mw2 after next/mw1 after next",
		},
		{
			title:   "build handler with multiple middleware (adding one middleware per Use())",
			handler: New().Use(middlewareOne).Use(middlewareTwo).Use(middlewareThree).Then(handlerFinal),
			out:     "/mw1 before next/mw2 before next/mw3 before next/final handler/mw3 after next/mw2 after next/mw1 after next",
		},
		{
			title:   "build handler with combination of single/plural calls of Use()",
			handler: New().Use(middlewareOne).Use(middlewareTwo, middlewareThree).Then(handlerFinal),
			out:     "/mw1 before next/mw2 before next/mw3 before next/final handler/mw3 after next/mw2 after next/mw1 after next",
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.handler.ServeHTTP(w, nil)
			if w.Body.String() != tc.out {
				t.Errorf("the output %q is expected to be %q", w.Body.String(), tc.out)
			}
		})
	}
}

func Test_Chain(t *testing.T) {
	// replace blob handler in order to check if it is being called
	blobHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/blob handler"))
	}

	type testCase struct {
		title string
		args  []interface{}
		out   string
		panic bool
	}

	cases := []testCase{
		{
			title: "building handler with unsupported argument types should panic",
			args: []interface{}{
				middlewareOne,
				middlewareTwo,
				true,
				middlewareThree,
				handlerFinal,
			},
			panic: true,
		},
		{
			title: "middleware should have control over the \"next\" handlers",
			args: []interface{}{
				middlewareOne,
				middlewareTwo,
				middlewareBreak,
				middlewareThree,
				handlerFinal,
			},
			out: "/mw1 before next/mw2 before next/skip the rest/mw2 after next/mw1 after next",
		},
		{
			title: "calling function without any arguments should build a middleware with only blobHandler",
			out:   "/blob handler",
		},
		{
			title: "building handler with all kind of supported arguments should be successful",
			args: []interface{}{
				middlewareOne,
				Middleware(middlewareTwo),
				middlewareFuncOne,
				MiddlewareFunc(middlewareFuncTwo),
				handlerOne,
				http.HandlerFunc(handlerTwo),
				middlewareThree,
				handlerFinal,
			},
			out: "/mw1 before next/mw2 before next/mw func1 before next/mw func2 before next" +
				"/first handler/second handler/mw3 before next/final handler/blob handler" +
				"/mw3 after next/mw func2 after next/mw func1 after next/mw2 after next/mw1 after next",
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					if tc.panic {
						t.Errorf("the code did not panic")
					}
				} else {
					if !tc.panic {
						t.Errorf("the code should not panic")
					}
				}
			}()
			w := httptest.NewRecorder()
			Chain(tc.args...).ServeHTTP(w, nil)
			if w.Body.String() != tc.out {
				t.Errorf("out %v expected to be %v", w.Body.String(), tc.out)
			}
		})
	}
}
