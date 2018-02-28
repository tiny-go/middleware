package mw

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	tokWrongIssuer = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJXcm9uZyIsImV4cCI6OTUwNzgwOTg3MiwiaXNzIjoiV3JvbmcifQ.fGVjxmejYo6J29fjGOXOFoh1r2k9oV0yfOOdBsYdacQ"
	tokValid       = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJBcHBJRCIsImV4cCI6OTUwNzgwOTg3MiwiaXNzIjoiQXBwSUQifQ.d-HgNTLWnKYhio9PkDp_tBt3oybFseNdirPaJgLudEw"
)

type invalidClaims struct{}

func (ic invalidClaims) Valid() error { return errors.New("error") }

func Test_JwtHS256(t *testing.T) {
	type testCase struct {
		title   string
		secret  string
		closure func() jwt.Claims
		headers map[string]string
		code    int
		body    string
	}

	cases := []testCase{
		{
			title:   "HTTP request without JWT should return an error",
			closure: func() jwt.Claims { return new(jwt.StandardClaims) },
			code:    http.StatusUnauthorized,
			body:    "no JSON web token in request\n",
		},
		{
			title:   "HTTP request conaining JWT with wrong number of segments should return an error",
			closure: func() jwt.Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{jwtAuthKey: "invalid token"},
			code:    http.StatusForbidden,
			body:    "token contains an invalid number of segments\n",
		},
		{
			title:   "claims validation failure should produce an error",
			secret:  "secret",
			closure: func() jwt.Claims { return new(invalidClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-IDcSemACt8x4iTMCda8Yhe3iZaWbvV5XKSTbuAn0M",
			},
			code: http.StatusForbidden,
			body: "error\n",
		},
		{
			title:   "HTTP request with expired token should return an error",
			secret:  "secret",
			closure: func() jwt.Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjF9.8onrqJhmsoas7S-2eOXSmQe1UZfbsK0zZyIw7ik8gZE",
			},
			code: http.StatusForbidden,
			body: "token is expired by",
		},
		{
			title:   "HTTP request with immortal token should end with success",
			secret:  "secret",
			closure: func() jwt.Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-IDcSemACt8x4iTMCda8Yhe3iZaWbvV5XKSTbuAn0M",
			},
			code: http.StatusOK,
		},
		{
			title:   "token with invalid signature should return an error",
			closure: func() jwt.Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-IDcSemACt8x4iTMCda8Yhe3iZaWbvV5XKSTbuAn0M",
			},
			code: http.StatusForbidden,
			body: "signature is invalid\n",
		},
	}

	t.Run("Given middleware which should parse JWT token signed with HMAC method", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.title, func(t *testing.T) {
				handler := Chain(JwtHS256(tc.secret, tc.closure), blobHandler)
				r, _ := http.NewRequest("", "", nil)
				w := httptest.NewRecorder()
				for hKey, hVal := range tc.headers {
					r.Header.Set(hKey, hVal)
				}
				handler.ServeHTTP(w, r)
				if w.Code != tc.code {
					t.Errorf("response status code %d was expected to be %d", w.Code, tc.code)
				}
				if !strings.Contains(w.Body.String(), tc.body) {
					t.Errorf("the response body %q was expected to contain substring %q, but didn't", w.Body.String(), tc.body)
				}
			})
		}
	})
}
