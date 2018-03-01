package mw

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	})
}

func Test_JwtHS256(t *testing.T) {
	type testCase struct {
		title     string
		secret    string
		closure   func() Claims
		headers   map[string]string
		code      int
		body      string
		hasClaims bool
	}

	cases := []testCase{
		{
			title:   "HTTP request without JWT should return an error",
			closure: func() Claims { return new(jwt.StandardClaims) },
			code:    http.StatusUnauthorized,
			body:    "no JSON web token in request\n",
		},
		{
			title:   "HTTP request conaining JWT with wrong number of segments should return an error",
			closure: func() Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{jwtAuthKey: "invalid token"},
			code:    http.StatusForbidden,
			body:    "token contains an invalid number of segments\n",
		},
		{
			title:   "claims validation failure should produce an error",
			secret:  "secret",
			closure: func() Claims { return new(invalidClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-IDcSemACt8x4iTMCda8Yhe3iZaWbvV5XKSTbuAn0M",
			},
			code: http.StatusForbidden,
			body: "error\n",
		},
		{
			title:   "HTTP request with expired token should return an error",
			secret:  "secret",
			closure: func() Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjF9.8onrqJhmsoas7S-2eOXSmQe1UZfbsK0zZyIw7ik8gZE",
			},
			code: http.StatusForbidden,
			body: "token is expired by",
		},
		{
			title:   "HTTP request with invalid signing method should produce an error",
			secret:  "secret",
			closure: func() Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.e30.b-gLBOyB62jzeiETcpDg4wgLa9EcJcEN5Dh4Hna5Uvs6wqGWRco1uIxdsQJRTvsWPq63A_ZM9g7rjs-SEORyty1DqWNeqaK3uaECr5n80dL_oKcWUhzCDJbC2W_v4_2jQz4lz5m12FH-_N19RRymA_GeKuZMyvH0MUlitVfnjlA",
			},
			code: http.StatusForbidden,
			body: "unexpected signing method: RS256\n",
		},
		{
			title:   "token with invalid signature should return an error",
			closure: func() Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-IDcSemACt8x4iTMCda8Yhe3iZaWbvV5XKSTbuAn0M",
			},
			code: http.StatusForbidden,
			body: "signature is invalid\n",
		},
		{
			title:   "HTTP request with immortal token should end with success",
			secret:  "secret",
			closure: func() Claims { return new(jwt.StandardClaims) },
			headers: map[string]string{
				jwtAuthKey: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-IDcSemACt8x4iTMCda8Yhe3iZaWbvV5XKSTbuAn0M",
			},
			code:      http.StatusOK,
			hasClaims: true,
		},
	}

	t.Run("Given middleware which should parse JWT token signed with HMAC method", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.title, func(t *testing.T) {
				handler := Chain(
					JwtHS256(tc.secret, tc.closure),
					func(w http.ResponseWriter, r *http.Request) {
						claims := GetClaimsFromContext(r.Context())
						if claims == nil && tc.hasClaims {
							t.Error("context was expected to contain claims but did not contain them")
						}
					},
				)
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
					t.Errorf("the response body %q was expected to contain substring %q, but did not contain it", w.Body.String(), tc.body)
				}
			})
		}
	})
}
