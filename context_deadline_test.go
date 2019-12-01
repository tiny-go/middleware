package mw

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func closureHandler100MSec(w http.ResponseWriter, r *http.Request) {
	var values []int
	err := Go(r.Context(), func(stop <-chan struct{}) error {
		for i := 0; i < 10; i++ {
			values = append(values, i)
			select {
			case <-stop:
				return nil
			case <-time.After(10 * time.Millisecond):
			}
		}
		return nil
	})
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(fmt.Sprintf("%v", values)))
	}
}

func handler100MSec(w http.ResponseWriter, r *http.Request) {
	var values []int
	for i := 0; i < 10; i++ {
		values = append(values, i)
		select {
		case <-r.Context().Done():
			return
		case <-time.After(10 * time.Millisecond):
		}
	}
	w.Write([]byte(fmt.Sprintf("%v", values)))
}

func Test_ContextDeadline(t *testing.T) {
	type testCase struct {
		title   string
		handler http.Handler
		out     string
	}

	cases := []testCase{
		{
			title:   "should return a proper message if context deadline exceeded (context handler)",
			handler: ContextDeadline(100 * time.Millisecond)(ContextHandler(http.HandlerFunc(handler100MSec))),
			out:     "context deadline exceeded\n",
		},
		{
			title:   "should return the result if it was obtained before context deadline (context handler)",
			handler: ContextDeadline(150 * time.Millisecond)(ContextHandler(http.HandlerFunc(handler100MSec))),
			out:     "[0 1 2 3 4 5 6 7 8 9]",
		},
		{
			title:   "should wait for the result if context does not have a deadline (context handler)",
			handler: ContextDeadline(0)(ContextHandler(http.HandlerFunc(handler100MSec))),
			out:     "[0 1 2 3 4 5 6 7 8 9]",
		},
		{
			title:   "should return a proper message if context deadline exceeded (closure handler)",
			handler: ContextDeadline(100 * time.Millisecond)(http.HandlerFunc(closureHandler100MSec)),
			out:     "context deadline exceeded",
		},
		{
			title:   "should return the result if it was obtained before context deadline (closure handler)",
			handler: ContextDeadline(150 * time.Millisecond)(http.HandlerFunc(closureHandler100MSec)),
			out:     "[0 1 2 3 4 5 6 7 8 9]",
		},
		{
			title:   "should wait for the result if context does not have a deadline (closure handler)",
			handler: ContextDeadline(0)(http.HandlerFunc(closureHandler100MSec)),
			out:     "[0 1 2 3 4 5 6 7 8 9]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("", "", nil)
			tc.handler.ServeHTTP(w, r)
			if w.Body.String() != tc.out {
				t.Errorf("the output %q is expected to be %q", w.Body.String(), tc.out)
			}
		})
	}
}
