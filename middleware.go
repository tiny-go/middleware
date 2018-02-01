package mw

import (
	"fmt"
	"net/http"
)

// Middleware in itself simply takes a http.Handler as the first parameter, wraps
// it and returns a new http.Handler for the server to call.
type Middleware func(http.Handler) http.Handler

// New is just an empty Midleware needed in order to start the chain:
// mw.New().Use(mwOne).Use(mwTwo, mwThree).Then(handler)
func New() Middleware {
	return func(handler http.Handler) http.Handler {
		return handler
	}
}

// Use transforms provided middleware function(s) (including current one) into
// a single middleware func.
func (mw Middleware) Use(middlewares ...Middleware) Middleware {
	for _, next := range middlewares {
		mw = func(curr, next Middleware) Middleware {
			return func(handler http.Handler) http.Handler {
				return curr(next(handler))
			}
		}(mw, next)
	}
	return mw
}

// Then injects handler into middleware chain.
func (mw Middleware) Then(final http.Handler) http.Handler {
	return mw(func(handler http.Handler) http.Handler {
		return handler
	}(final))
}

// blobHandler will be called anyway (if request reached your final handler),
// it is just a blob used to prevent calling ServeHTTP() method from panic since
// final handler is being wrapped with anonymous Middleware func. It has been
// declared as a variable intentionally - to be able to define a custom handler.
var blobHandler = func(w http.ResponseWriter, r *http.Request) {}

// Chain builds a http.Handler from passed arguments. It accepts different
// kinds of argument types:
// - Middleware - can break the chain inside
// - func(http.Handler) http.Handler - same with Middleware
// - http.Handler - next will be called in any case
// - func(w http.ResponseWriter, r *http.Request) - sme with http.Handler
// Keep in mind:
// - by passing http.Handler/http.HandlerFunc instead of Middleware you lose
// control over the next Middleware (no way to cancel it), but if you only need
// to put something to the context (and do not have any logic after calling "next")
// there is no sense to build Middleware func around
// - even if you do not pass any handlers blobHandler will be executed.
func Chain(handlers ...interface{}) http.Handler {
	// fake handler in order to wrap last handler call "next"
	var f http.Handler = http.HandlerFunc(blobHandler)
	// apply middleware/handlers from the last to the first one
	for i := range handlers {
		switch t := handlers[len(handlers)-1-i].(type) {
		// wrap existing handler (or blobHandler) with a func
		case func(http.Handler) http.Handler:
			f = t(f)
		// wrap existing handler (or blobHandler) with a Middleware
		case Middleware:
			f = t(f)
		// ordinary functions can also be provided as arguments, in such case they
		// will be called via adapter http.HandlerFunc
		case func(w http.ResponseWriter, r *http.Request):
			f = func(curr, next http.Handler) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					curr.ServeHTTP(w, r)
					// due to the blobHandler next will never be nil
					next.ServeHTTP(w, r)
				}
			}(http.HandlerFunc(t), f)
		// since http.HandlerFunc implements http.Handler interface we can use type
		// http.Handler for both of them
		case http.Handler:
			f = func(curr, next http.Handler) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					curr.ServeHTTP(w, r)
					next.ServeHTTP(w, r)
				}
			}(t, f)
		default:
			// everything else is not supported
			panic(fmt.Sprintf("unsupported argument type \"%T\"", t))
		}
	}
	return f
}
