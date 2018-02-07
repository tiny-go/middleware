package mw

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func do(ctx context.Context, handler func(stop <-chan struct{}) error) error {
	// error chan
	errChan := make(chan error, 1)
	stopChan := make(chan struct{}, 1)
	// call handler in goroutine
	go func() { errChan <- handler(stopChan) }()
	// wait until context deadline or job is done
	select {
	// job was done
	case err := <-errChan:
		return err
	// timeout
	case <-ctx.Done():
		// send stop signal to the handler
		stopChan <- struct{}{}
		return ctx.Err()
	}
}

func handler100MSec(w http.ResponseWriter, r *http.Request) {
	var values []int
	err := do(r.Context(), func(stop <-chan struct{}) error {
		for i := 0; i < 10; i++ {
			values = append(values, i)
			select {
			case <-stop:
				return nil
			default:
				time.Sleep(10 * time.Millisecond)
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

func Test_ContextDeadline(t *testing.T) {
	type testCase struct {
		title   string
		handler http.Handler
		out     string
	}

	cases := []testCase{
		{
			title:   "should return a proper message if context deadline exceeded",
			handler: ContextDeadline(100 * time.Millisecond)(http.HandlerFunc(handler100MSec)),
			out:     "context deadline exceeded",
		},
		{
			title:   "should return the result if it was obtained before context deadline",
			handler: ContextDeadline(150 * time.Millisecond)(http.HandlerFunc(handler100MSec)),
			out:     "[0 1 2 3 4 5 6 7 8 9]",
		},
		{
			title:   "should wait for the result if context does not have a deadline",
			handler: ContextDeadline(0)(http.HandlerFunc(handler100MSec)),
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
