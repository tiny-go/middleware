package mw

import (
	"encoding/json"
	"errors"
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
				case <-time.After(10 * time.Millisecond):
				}
			}
			// do not forget to complete the task (otherwise it will stay "in progress" forever)
			return job.Complete(reply, nil)
		})
	}
	// return current task to be processed by handleResponse func
	return job
}

func Test_task_Complete(t *testing.T) {
	t.Run("should throw an error trying to complete the task which was already done", func(t *testing.T) {
		job := &task{status: StatusDone}
		if err := job.Complete(nil, nil); err != ErrAlreadyDone {
			t.Errorf("should return error: %s", ErrAlreadyDone)
		}
	})
	t.Run("should throw an error trying to complete the task which was not started", func(t *testing.T) {
		job := &task{status: StatusWaiting}
		if err := job.Complete(nil, nil); err != ErrNotStarted {
			t.Errorf("should return error: %s", ErrNotStarted)
		}
	})
	t.Run("should pass actual error trying to complete the task which is in progress", func(t *testing.T) {
		job := &task{status: StatusInProgress}
		actual := errors.New("Actual error")
		if err := job.Complete(nil, actual); err != actual {
			t.Errorf("should return error: %s", actual)
		}
	})
}

func Test_AsyncRequest_input_arguments(t *testing.T) {
	t.Run("request timeout should be less than async timeout", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("the code should panic")
			}
		}()
		AsyncRequest(100, 50, 80)(http.HandlerFunc(blobHandler))
	})
	t.Run("keep result should be greater than async timeout", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("the code should panic")
			}
		}()
		AsyncRequest(100, 200, 150)(http.HandlerFunc(blobHandler))
	})
}

func Test_AsyncRequest(t *testing.T) {
	type request struct {
		title      string
		timeout    time.Duration
		timestamps bool
		hasID      bool
		code       int
		data       string
	}

	type testCase struct {
		title    string
		headers  map[string]string
		handler  http.Handler
		requests []request
	}

	cases := []testCase{
		{
			title:   "async middleware with synchronous request",
			headers: make(map[string]string), // just in case to avoid panics
			handler: AsyncRequest(10*time.Millisecond, 20*time.Millisecond, 30*time.Millisecond)(
				handleResponse(handlerAsync),
			),
			requests: []request{
				{
					title: "should fail with timeout error if handler does not have enough time to complete the task",
					code:  http.StatusRequestTimeout,
					data:  "context deadline exceeded\n",
				},
			},
		},
		{
			title:   "async middleware with synchronous request",
			headers: make(map[string]string), // just in case to avoid panics
			handler: AsyncRequest(200*time.Millisecond, 300*time.Millisecond, 500*time.Millisecond)(
				handleResponse(handlerAsync),
			),
			requests: []request{
				{
					title: "should be successful if handler has enough time to complete the task",
					code:  http.StatusOK,
					data:  "[0,1,2,3,4,5,6,7,8,9]\n",
				},
			},
		},
		{
			title: "async middleware with asynchronous request",
			headers: map[string]string{
				asyncHeader: "",
			},
			handler: AsyncRequest(50*time.Millisecond, 300*time.Millisecond, 500*time.Millisecond)(
				handleResponse(handlerAsync),
			),
			requests: []request{
				{
					title:      "should produce response with request ID if handler did not have enough time to complete the task",
					hasID:      true,
					timestamps: true,
					code:       http.StatusAccepted,
					data:       "request is in progress\n",
				},
				{
					title:      "should provide status of the current job if async request is still in progress",
					timeout:    20 * time.Millisecond,
					hasID:      true,
					timestamps: true,
					code:       http.StatusAccepted,
					data:       "request is in progress\n",
				},
				{
					title:   "should store the result after task is completed and be able to return it (in cooperation with handler)",
					timeout: 50 * time.Millisecond,
					code:    http.StatusOK,
					data:    "[0,1,2,3,4,5,6,7,8,9]\n",
				},
				{
					title: "should provided the result only once and delete after that",
					code:  http.StatusBadRequest,
					data:  "invalid or expired request\n",
				},
			},
		},
		{
			title: "async middleware with asynchronous request",
			headers: map[string]string{
				asyncHeader: "",
			},
			handler: AsyncRequest(200*time.Millisecond, 300*time.Millisecond, 500*time.Millisecond)(
				handleResponse(handlerAsync),
			),
			requests: []request{
				{
					title: "should be successful if handler has enough time to complete the task",
					code:  http.StatusOK,
					data:  "[0,1,2,3,4,5,6,7,8,9]\n",
				},
			},
		},
		{
			title: "async middleware with asynchronous request",
			headers: map[string]string{
				asyncHeader: "",
			},
			handler: AsyncRequest(50*time.Millisecond, 199*time.Millisecond, 200*time.Millisecond)(
				handleResponse(handlerAsync),
			),
			requests: []request{
				{
					title: "should produce response with request ID if handler did not have enough time to complete the task",
					hasID: true,
					code:  http.StatusAccepted,
					data:  "request is in progress\n",
				},
				{
					title:   "should be deleted (as expired) after keep result timeout",
					timeout: 300 * time.Millisecond,
					code:    http.StatusBadRequest,
					data:    "invalid or expired request\n",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			for _, req := range tc.requests {
				t.Run(req.title, func(t *testing.T) {
					// sleep before request
					time.Sleep(req.timeout)
					w := httptest.NewRecorder()
					r, _ := http.NewRequest("", "", nil)
					// copy headers
					for key, value := range tc.headers {
						r.Header.Set(key, value)
					}
					tc.handler.ServeHTTP(w, r)
					// compare status code
					if w.Code != req.code {
						t.Errorf("status code %d is expected to be %d", w.Code, req.code)
					}
					// compare response body
					if w.Body.String() != req.data {
						t.Errorf("the output %q is expected to be %q", w.Body.String(), req.data)
					}
					// set request ID (for the next async requests)
					if id := w.Header().Get(asyncRequestID); id != "" {
						if !req.hasID {
							t.Error("the response should not contain request id")
						}
						tc.headers[asyncRequestID] = id
					} else if req.hasID {
						t.Error("the response should contain request id")
					}
					if req.timestamps {
						started, keepUntil := w.Header().Get(asyncRequestAccepted), w.Header().Get(asyncRequestKeepUntil)
						if _, err := time.Parse(DefaultTimeFormat, started); err != nil {
							t.Errorf("should contain %q header and have valid format %q", asyncRequestAccepted, DefaultTimeFormat)
						}
						if _, err := time.Parse(DefaultTimeFormat, keepUntil); err != nil {
							t.Errorf("should contain %q header and have valid format %q", asyncRequestKeepUntil, DefaultTimeFormat)
						}
					}
				})
			}
		})
	}
}
