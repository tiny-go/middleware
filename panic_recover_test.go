package mw

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tiny-go/errors"
)

func Test_PanicRecover(t *testing.T) {
	type testCase struct {
		title       string
		nextHandler http.Handler
		status      int
		output      string
	}

	cases := []testCase{
		{
			title: "ignore \"nil\" panic and send a success code",
			nextHandler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					panic(nil)
				},
			),
			status: http.StatusOK,
		},
		{
			title: "catch \"unknown\" panic and report to the client with code 500",
			nextHandler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					panic("unexpected panic")
				},
			),
			status: http.StatusInternalServerError,
			output: "unexpected panic\n",
		},
		{
			title: "catch a \"standard error\" and report to the client with code 500",
			nextHandler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					panic(fmt.Errorf("standard error"))
				},
			),
			status: http.StatusInternalServerError,
			output: "standard error\n",
		},
		{
			title: "catch an \"error with status code\" and report to the client",
			nextHandler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					panic(errors.NewBadRequest("bad request"))
				},
			),
			status: http.StatusBadRequest,
			output: "bad request\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Error("the code should never panic because it is wrapped with PanicRecover middleware")
				}
			}()
			w := httptest.NewRecorder()
			PanicRecover(errors.Send)(tc.nextHandler).ServeHTTP(w, nil)
			if w.Code != tc.status {
				t.Errorf("status code %d was expected to be %d", w.Code, tc.status)
			}
			if w.Body.String() != tc.output {
				t.Errorf("net output %q was expected to be %q", w.Body.String(), tc.output)
			}
		})
	}
}
