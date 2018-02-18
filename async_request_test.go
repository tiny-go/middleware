package mw

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func handleResponse(handler func(w http.ResponseWriter, r *http.Request) HandlerTask) http.HandlerFunc {
	// create wrapper handler
	return func(w http.ResponseWriter, r *http.Request) {
		// call handler
		job := handler(w, r)
		// read the result
		data, err := job.Resolve()
		// check error
		switch err {
		// job was successfully done
		case nil:
			// send the response
			json.NewEncoder(w).Encode(data)
		// do nothing: middleware will send the response depending on request type
		case ErrNotCompleted:
			// return or ignore
			return
		// done with error
		default:
			// send error
			http.Error(w, err.Error(), 500)
		}
	}
}

func handlerAsync(w http.ResponseWriter, r *http.Request) HandlerTask {
	// get job from context
	job, _ := GetHandlerTask(r.Context())
	// start the job if it is new one, otherwise pass the job to handleResponse func
	if job.Status() == StatusWaiting {
		// define reply type
		var reply []int
		// start the job
		job.Do(r.Context(), func(stop <-chan struct{}) error {
			// emulate task which takes a lot of time to complete
			for i := 0; i < 10; i++ {
				// add values one by one
				reply = append(reply, i)
				// catch stop signal or wait
				select {
				// request timeout (context deadline - stopped externally)
				case <-stop:
					// do something to terminate handler
					return nil
				// wait
				default:
					time.Sleep(10 * time.Millisecond)
				}
			}
			// do not forget to complete the task (otherwise it will stay "in progress" forever)
			return job.Complete(reply, nil)
		})
	}
	// return current task to be processed by handleResponse func
	return job
}

func Test_AsyncRequest(t *testing.T) {
	type request struct {
		title   string
		handler http.Handler
		headers http.Header
		timeout time.Duration
		code    int
		data    string
	}

	type testCase struct {
		title    string
		requests []request
	}

	cases := []testCase{
		{
			title: "async middleware with synchronous call",
			requests: []request{
				{
					title: "sync request should fail with timeout error if handler does not have enough time to complete the task",
					handler: AsyncRequest(10*time.Millisecond, 20*time.Millisecond, 30*time.Millisecond)(
						handleResponse(handlerAsync),
					),
					code: http.StatusRequestTimeout,
					data: "task has not been completed\n",
				},
				{
					title: "sync request should be successful if handler has enough time to complete the task",
					handler: AsyncRequest(200*time.Millisecond, 300*time.Millisecond, 500*time.Millisecond)(
						handleResponse(handlerAsync),
					),
					code: http.StatusOK,
					data: "[0,1,2,3,4,5,6,7,8,9]\n",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {

			log.Println(asyncJobs)

			for _, req := range tc.requests {
				t.Run(req.title, func(t *testing.T) {
					// sleep before request
					time.Sleep(req.timeout)
					w := httptest.NewRecorder()
					r, _ := http.NewRequest("", "", nil)
					req.handler.ServeHTTP(w, r)
					// compare status code
					if w.Code != req.code {
						t.Errorf("status code %d is expected to be %d", w.Code, req.code)
					}
					// compare response body
					if w.Body.String() != req.data {
						t.Errorf("the output %q is expected to be %q", w.Body.String(), req.data)
					}
				})
			}
		})
	}
}
