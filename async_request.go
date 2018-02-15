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
	// AsyncWaiting is initial status - job is not started yet
	AsyncWaiting AsyncStatus = iota
	// AsyncInProgress indicates that job is started but not finished yet
	AsyncInProgress
	// AsyncDone task is done
	AsyncDone
)

// AsyncStatus represents the status of asynchronous task.
type AsyncStatus int

const (
	asyncHeader                = "Async-Request"
	asyncRequestID             = "Async-Request-ID"
	asyncContextKey contextKey = "async-request"
)

var asyncJobs = map[string]*Async{}

// TODO: async pool (watcher.Add(task)) with watcher which can finish and remove expired tasks

// TODO:
// Task ...
type Task interface {
}

// TODO:
// Sync ...
type Sync struct {
}

// Async represents asynchronous handler job.
type Async struct {
	ID string

	Data  interface{}
	Error error

	status   AsyncStatus
	started  time.Time
	finished time.Time
	// TODO: use it!!!
	asyncTimeout time.Duration
}

// NewAsync ...
func NewAsync() *Async {
	id := md5.Sum([]byte(time.Now().String()))
	return &Async{
		ID: hex.EncodeToString(id[:]),
	}
}

// Status returns status of the current task.
func (t *Async) Status() AsyncStatus {
	return t.status
}

// Complete the task.
func (t *Async) Complete(err error) error {
	switch t.status {
	case AsyncWaiting:
		return errors.New("job has not been started")
	case AsyncDone:
		return errors.New("job already completed")
	default:
		t.Error, t.status, t.finished = err, AsyncDone, time.Now()
		return nil
	}
}

// Do ...
func (t *Async) Do(ctx context.Context, handler func(stop <-chan struct{}) error) error {
	// memorize start time and change job status
	t.status, t.started = AsyncInProgress, time.Now()
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
							async = NewAsync()
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
							if async.status == AsyncDone {
								delete(asyncJobs, async.ID)
							} else {
								// return request ID
								w.Header().Set(asyncRequestID, async.ID)
								// the status ot request is "accepted"
								w.WriteHeader(http.StatusAccepted)
								// provide a basic info message to the client
								w.Write([]byte("request is still in progress"))
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

// GetAsyncFromContext ...
func GetAsyncFromContext(ctx context.Context) (*Async, bool) {
	async, ok := ctx.Value(asyncContextKey).(*Async)
	return async, ok
}
