package mw

import (
	"encoding/json"
	"net/http"
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
		var reply struct{ Numbers []int }
		// start the job
		job.Do(r.Context(), func(stop <-chan struct{}) error {
			// emulate task which takes a lot of time to complete
			for i := 0; i < 100; i++ {
				// add values one by one
				reply.Numbers = append(reply.Numbers, i)
				// catch stop signal or wait
				select {
				// request timeout (context deadline - stopped externally)
				case <-stop:
					// do something to terminate handler
					return nil
				// wait
				default:
					time.Sleep(time.Millisecond)
				}
			}
			// do not forget to complete the task (otherwise it will stay "in progress"
			// forever)
			return job.Complete(reply, nil)
		})
	}
	// return current task to be processed by handleResponse func
	return job
}

func Test_AsyncRequest(t *testing.T) {

}
