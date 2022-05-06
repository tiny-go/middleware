package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func Test_RequestLimiter(t *testing.T) {
	t.Run("process request when", func(t *testing.T) {
		var (
			wg    sync.WaitGroup
			limit = 10
			done  = make(chan struct{})
		)
		// use wait group to make sure that handlers were started
		wg.Add(limit)
		// close the channel on exit to unblock handlers
		defer close(done)

		mw := RequestLimiter(nil, limit)

		handler := mw(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				wg.Done()
				<-done
			},
		))

		for i := 0; i < limit; i++ {
			go handler.ServeHTTP(nil, nil)
		}

		wg.Wait()

		w := httptest.NewRecorder()
		mw(nil).ServeHTTP(w, nil)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("status code %d was expected to be %d", w.Code, http.StatusTooManyRequests)
		}

		expected := http.StatusText(http.StatusTooManyRequests) + "\n"
		if w.Body.String() != expected {
			t.Errorf("net output %q was expected to be %q", w.Body.String(), expected)
		}
	})
}
