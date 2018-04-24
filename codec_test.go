package mw

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tiny-go/codec/driver"
	"github.com/tiny-go/codec/driver/json"
	"github.com/tiny-go/codec/driver/xml"
)

func TestCodecFromList(t *testing.T) {
	type testCase struct {
		title   string
		handler http.Handler
		request *http.Request
		ctype   string
		code    int
		body    string
	}

	cases := []testCase{
		{
			title: "should throw an error if request codec in not supported",
			handler: PanicRecover(PanicHandler)(
				CodecFromList(driver.DummyRegistry{&json.JSON{}, &xml.XML{}})(nil),
			),
			request: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "", nil)
				r.Header.Set(contentTypeHeader, "unknown")
				return r
			}(),
			code: http.StatusBadRequest,
			body: "unsupported request codec: \"unknown\"\n",
		},
		{
			title: "should throw an error if response codec in not supported",
			handler: PanicRecover(PanicHandler)(
				CodecFromList(driver.DummyRegistry{&json.JSON{}, &xml.XML{}})(nil),
			),
			request: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "", nil)
				r.Header.Set(contentTypeHeader, "application/json")
				r.Header.Set(acceptHeader, "unknown")
				return r
			}(),
			code: http.StatusBadRequest,
			body: "unsupported response codec: \"unknown\"\n",
		},
		{
			title: "should find corresponding codecs and handle the request successfully",
			handler: PanicRecover(PanicHandler)(
				CodecFromList(driver.DummyRegistry{&json.JSON{}, &xml.XML{}})(
					BodyClose(
						http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							type Data struct{ Test string }
							var data Data
							RequestCodecFromContext(r.Context()).Decoder(r.Body).Decode(&data)
							ResponseCodecFromContext(r.Context()).Encoder(w).Encode(data)
						}),
					),
				),
			),
			request: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "", strings.NewReader("{\"test\":\"passed\"}\n"))
				r.Header.Set(contentTypeHeader, "application/json")
				r.Header.Set(acceptHeader, "application/xml")
				return r
			}(),
			code: http.StatusOK,
			body: "<Data><Test>passed</Test></Data>",
		},
	}

	t.Run("Given middleware function", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.title, func(t *testing.T) {
				w := httptest.NewRecorder()
				tc.handler.ServeHTTP(w, tc.request)

				if w.Code != tc.code {
					t.Errorf("status code %d was expected to be %d", w.Code, tc.code)
				}
				if w.Body.String() != tc.body {
					t.Errorf("response body %q was expected to be %q", w.Body.String(), tc.body)
				}
			})
		}
	})
}
