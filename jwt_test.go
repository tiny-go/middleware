package mw

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

type invalidClaims struct{}

func (ic invalidClaims) Valid() error { return errors.New("error") }

func Test_Bearer(t *testing.T) {
	tokenString := "this-is-a-token"
	t.Run("Given a function that should retrieve JWT from request", func(t *testing.T) {
		t.Run("get a token from the request headers", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "", nil)
			r.Header.Set(jwtAuthKey, tokenString)
			token, ok := Bearer(r)
			if !ok {
				t.Error("cannot retrieve token from request headers")
			}
			if token != tokenString {
				t.Error("token contains invalid string")
			}
		})
		t.Run("get a token from the request URI", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodDelete, "?"+jwtAuthKey+"="+tokenString, nil)
			token, ok := Bearer(r)
			if !ok {
				t.Error("cannot retrieve token from request URI")
			}
			if token != tokenString {
				t.Error("token contains invalid string")
			}
		})
		t.Run("getting a token from multipart form is not supported", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "", nil)
			r.Header.Add("Content-Type", "multipart/form-data")
			_, ok := Bearer(r)
			if ok {
				t.Error("should not be able to retrieve the token")
			}
		})
		t.Run("get a token from the request body sent as form-data", func(t *testing.T) {
			form := url.Values{}
			form.Add(jwtAuthKey, tokenString)
			r, _ := http.NewRequest(http.MethodPost, "", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			token, ok := Bearer(r)
			if !ok {
				t.Error("cannot retrieve token from request body")
			}
			if token != tokenString {
				t.Errorf("token contains invalid string: %q", token)
			}
		})
		t.Run("unable to get a token", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "", nil)
			token, ok := Bearer(r)
			if ok || token != "" {
				t.Error("request should not contain a token")
			}
		})
	})
}

func Test_ClaimsFromContextTo(t *testing.T) {
	t.Run("Given request context with claims object inside", func(t *testing.T) {
		claims := &jwt.StandardClaims{}
		ctx := context.WithValue(context.Background(), claimsKey{}, claims)
		t.Run("assign claims to a valid receiver", func(t *testing.T) {
			var recv *jwt.StandardClaims
			if err := ClaimsFromContextTo(ctx, &recv); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
		t.Run("no claims in the context", func(t *testing.T) {
			if err := ClaimsFromContextTo(context.Background(), nil); !reflect.DeepEqual(err, errors.New("no claims in the context")) {
				t.Errorf("expected error was not returned")
			}
		})
		t.Run("assign claims to an invalid receiver", func(t *testing.T) {
			var recv struct{}
			if err := ClaimsFromContextTo(ctx, &recv); !reflect.DeepEqual(err, errors.New("cannot assign claims to the provided receiver")) {
				t.Errorf("expected error was not returned")
			}
		})
		t.Run("assign claims to not a pointer", func(t *testing.T) {
			var recv *jwt.StandardClaims
			if err := ClaimsFromContextTo(ctx, recv); !reflect.DeepEqual(err, errors.New("claims object is not a pointer")) {
				t.Errorf("expected error was not returned")
			}
		})
	})
}
