package mw

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	// StatusWaiting is initial status - job is not started yet
	StatusWaiting JobStatus = iota
	// StatusInProgress indicates that job is started but not finished yet
	StatusInProgress
	// StatusDone task is done
	StatusDone
)

// JobStatus represents the status of asynchronous task.
type JobStatus int

const (
	asyncHeader                = "Async-Request"
	asyncRequestID             = "Async-Request-ID"
	asyncContextKey contextKey = "async-request"
)

// TODO: sync.Map
var asyncJobs = map[string]*Async{}

// TODO: async pool (watcher.Add(task)) with watcher which can finish and remove expired tasks

// TODO:
// Task ...
type HandlerTask interface {
	Do(context.Context, func(<-chan struct{}) error) error
	Status() JobStatus
	Resolve() (interface{}, error)
	Complete(interface{}, error) error
}

// TODO:
// Sync ...
type Sync struct {
}

// Async represents asynchronous handler job.
type Async struct {
	ID string

	data  interface{}
	error error

	status   JobStatus
	started  time.Time
	finished time.Time
	// TODO: use it!!!
	asyncTimeout time.Duration
}

// newAsync ... TODO: private
func newAsync() *Async {
	id := md5.Sum([]byte(time.Now().String()))
	return &Async{
		ID: hex.EncodeToString(id[:]),
	}
}

// Status returns status of the current task.
func (t *Async) Status() JobStatus {
	return t.status
}

// Resolve ...
func (t *Async) Resolve() (interface{}, error) {
	return t.data, t.error
}

// Complete the task.
func (t *Async) Complete(data interface{}, err error) error {
	switch t.status {
	case StatusWaiting:
		return errors.New("job has not been started")
	case StatusDone:
		return errors.New("job already completed")
	default:
		t.data, t.error, t.status, t.finished = data, err, StatusDone, time.Now()
		return nil
	}
}

// Do ...
func (t *Async) Do(ctx context.Context, handler func(stop <-chan struct{}) error) error {
	// memorize start time and change job status
	t.status, t.started = StatusInProgress, time.Now()
	// error chan
	errChan := make(chan error, 1)
	// call handler in goroutine TODO: inject new context with new ASYNC deadline here
	go func() { errChan <- handler(ctx.Done()) }()
	// wait until context deadline or job is done
	select {
	// job was done
	case err := <-errChan:
		return err
	// timeout
	case <-ctx.Done():
		// request timeout (but this is not an error and handler probably is still running)
		return nil
	}
}

// TODO: come up with sync/async timeout params for request and process

// AsyncRequest middleware provides a mechanism to request the data again after timeout.
func AsyncRequest(asyncTimeout time.Duration) Middleware {
	// create a new Middleware
	return func(next http.Handler) http.Handler {
		// define the httprouter.Handle
		return ContextDeadline(asyncTimeout)(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {

					log.Println(asyncJobs)

					// check if async request, if not - ignore the next code block
					if _, ok := r.Header[asyncHeader]; ok {
						// current request
						var async *Async
						// if contains ID - it is not a new request
						if requestID := r.Header.Get(asyncRequestID); requestID != "" {
							var ok bool
							// find async job
							if async, ok = asyncJobs[requestID]; !ok {
								// async request is expired or has invalid ID
								http.Error(w, "invalid or expired request", http.StatusBadRequest)
								// skip next handlers
								return
							}
						} else {
							// create new async task
							async = newAsync()
							// store job in the list
							asyncJobs[async.ID] = async
						}
						// get context from request
						ctx := r.Context()
						// put async task to the context
						ctx = context.WithValue(ctx, asyncContextKey, async)
						// replace request
						r = r.WithContext(ctx)
						// check async on exit and remove if it's done
						defer func() {
							if async.status == StatusDone {
								delete(asyncJobs, async.ID)
							} else {
								// return request ID
								w.Header().Set(asyncRequestID, async.ID)
								// the status ot request is "accepted"
								w.WriteHeader(http.StatusAccepted)
								// provide a basic info message to the client
								w.Write([]byte("request is in progress"))
							}
						}()
					}
					// call next handler
					next.ServeHTTP(w, r)
				},
			),
		)
	}
}

// GetHandlerTask ...
func GetHandlerTask(ctx context.Context) (HandlerTask, bool) {
	async, ok := ctx.Value(asyncContextKey).(HandlerTask)
	return async, ok
}
