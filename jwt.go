package middleware

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/tiny-go/errors"
)

// jwtAuthKey is an authorization key param.
const jwtAuthKey = "Authorization"

// claimsKey represents claims context key that never collides.
type claimsKey struct{}

// Claims interface.
type Claims interface {
	Valid() error
}

// JWTParser interface is responsible for token validation, meanwhile it parses
// token to the provided receiver.
type JWTParser interface {
	Parse(jwt string, recv *Claims) error
}

// ClaimsFactory is a func that returns a new custom claims when called.
type ClaimsFactory func() Claims

// JWT is a JSON Web token middleware that parses token with provided parser
// to the provided Claims receiver and puts it to the request context.
func JWT(parser JWTParser, cf ClaimsFactory) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get JSON web token from the request
			bearer, ok := Bearer(r)
			if !ok {
				http.Error(w, "no JSON web token in request", http.StatusUnauthorized)
				return
			}
			// instantiate an empty claims
			claims := cf()
			// validate token
			if err := parser.Parse(bearer, &claims); err != nil {
				errors.Send(w, errors.NewUnauthorized(err))
				return
			}
			// add claims to the context and call the next
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims)))
		})
	}
}

// Bearer gets the bearer out of a given request object.
func Bearer(r *http.Request) (string, bool) {
	var bearer string
	// get bearer from the request headers
	if bearer = r.Header.Get(jwtAuthKey); bearer != "" {
		return bearer, true
	}
	// try URL params for the bearer
	for _, bearer = range r.URL.Query()[jwtAuthKey] {
		if bearer != "" {
			return bearer, true
		}
	}
	// try to parse headers from request body
	if err := r.ParseForm(); err != nil {
		return "", false
	}
	if bearer = r.FormValue(jwtAuthKey); bearer != "" {
		// remove JWT from the form values
		delete(r.PostForm, jwtAuthKey)
		return bearer, true
	}
	// token not found
	return "", false
}

// ClaimsFromContextTo retrieves claims from context and assigns to the provided receiver.
func ClaimsFromContextTo(ctx context.Context, recv interface{}) (err error) {
	// check context
	claims := ctx.Value(claimsKey{})
	if claims == nil {
		return fmt.Errorf("no claims in the context")
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot assign claims to the provided receiver")
		}
	}()
	// check receiver type (should be a pointer)
	rv := reflect.ValueOf(recv)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("claims object is not a pointer")
	}
	// assign context claims to the provided receiver
	rv.Elem().Set(reflect.ValueOf(claims))
	// success (if no panic in the previous line)
	return nil
}
