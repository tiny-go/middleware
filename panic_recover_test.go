package mw

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_PanicRecover(t *testing.T) {
	type testCase struct {
		title       string
		nextHandler http.Handler
		logOutput   string
		netStatus   int
		netOutput   string
	}

	cases := []testCase{
		{
			title: "catch \"unknown\" panic and report to the log",
			nextHandler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) { panic("unexpected panic") },
			),
			netStatus: http.StatusInternalServerError,
			logOutput: "Recovered from panic:it should panic",
			netOutput: "unexpected panic\n",
		},
		{
			title: "catch a \"standard error\" and report to the client with code 500",
			nextHandler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) { panic(errors.New("standard error")) },
			),
			netStatus: http.StatusInternalServerError,
			netOutput: "standard error\n",
		},
		{
			title: "catch an \"error with status code\" and report to the client",
			nextHandler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					panic(NewStatusError(http.StatusBadRequest, errors.New("bad request")))
				},
			),
			netStatus: http.StatusBadRequest,
			netOutput: "bad request\n",
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
			PanicRecover(PanicHandler)(tc.nextHandler).ServeHTTP(w, nil)
			if w.Code != tc.netStatus {
				t.Errorf("status code %d was expected to be %d", w.Code, tc.netStatus)
			}
			if w.Body.String() != tc.netOutput {
				t.Errorf("net output %q was expected to be %q", w.Body.String(), tc.netOutput)
			}
		})
	}
}
