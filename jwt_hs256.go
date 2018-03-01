package mw

import (
	"context"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// claimsKey represents claims context key.
	claimsKey contextKey = "request-claims"
	// jwtAuthKey is an authorization key param.
	jwtAuthKey = "Authorization"
)

// Claims is a vrapper over jwt.Claims to avoid import of jwt-go using JwtHS256
// middleware.
type Claims interface {
	jwt.Claims
}

// JwtHS256 is a JSON Web token middleware using HMAC signing method.
func JwtHS256(secret string, cf func() Claims) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get JSON web token from the request
			bearer, ok := Bearer(r)
			if !ok {
				http.Error(w, "no JSON web token in request", http.StatusUnauthorized)
				return
			}
			// parse JWT
			token, err := jwt.ParseWithClaims(bearer, cf(), func(token *jwt.Token) (interface{}, error) {
				// check token algorithm
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})
			// cannot parse the token
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			// token validation error
			if !token.Valid {
				http.Error(w, "token is invalid", http.StatusForbidden)
				return
			}
			// add claims to the context and call next
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, token.Claims)))
		})
	}
}

// Bearer gets the bearer out of given request.
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

// GetClaimsFromContext returns claims from context.
func GetClaimsFromContext(ctx context.Context) jwt.Claims {
	claims, _ := ctx.Value(claimsKey).(jwt.Claims)
	return claims
}
