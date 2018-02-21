package mw

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"

	timap "github.com/Alma-media/go-timap"
)

const (
	// StatusWaiting is initial status - job is not started yet.
	StatusWaiting JobStatus = iota
	// StatusInProgress indicates that job is started but not finished yet.
	StatusInProgress
	// StatusDone indicates that task was done.
	StatusDone
)

// JobStatus represents the status of asynchronous task.
type JobStatus int

const (
	asyncHeader                = "Async-Request"
	asyncRequestID             = "Async-Request-ID"
	asyncContextKey contextKey = "async-request"
)

var (
	// ErrNotCompleted - current job was not completed.
	ErrNotCompleted = errors.New("task has not been completed")
	// ErrNotStarted - current job was not started.
	ErrNotStarted = errors.New("job has not been started")
	// ErrAlreadyDone - current job has been already done.
	ErrAlreadyDone = errors.New("job already completed")
)

// HandlerTask represents sync/async handler task.
type HandlerTask interface {
	// Do should execute provided closure.
	Do(context.Context, func(<-chan struct{}) error)
	// Status should return the status of current job/task.
	Status() JobStatus
	// Complete is supposed to be called inside Do's closure when job is done in
	// order to change job status and be able to return the result by calling
	// Resolve() func.
	Complete(interface{}, error) error
	// Resolve returns the result (which should be returned from the clousere using
	// Complete()) and error.
	Resolve() (interface{}, error)
}

// base task for sync/async jobs.
type task struct {
	sync.Mutex
	status       JobStatus
	started      time.Time
	finished     time.Time
	asyncTimeout time.Duration
	// returning params
	data  interface{}
	error error
}

// Status returns status of the current task.
func (t *task) Status() JobStatus {
	t.Lock()
	defer t.Unlock()
	return t.status
}

// Resolve returns the result of handler execution and an error.
func (t *task) Resolve() (interface{}, error) {
	t.Lock()
	defer t.Unlock()
	if t.status != StatusDone {
		return nil, ErrNotCompleted
	}
	return t.data, t.error
}

// Complete the task with some result and error, change status to "done".
func (t *task) Complete(data interface{}, err error) error {
	t.Lock()
	defer t.Unlock()
	switch t.status {
	case StatusWaiting:
		return ErrNotStarted
	case StatusDone:
		return ErrAlreadyDone
	default:
		t.data, t.error, t.status, t.finished = data, err, StatusDone, time.Now()
		return err
	}
}

// syncTask represents synchronous handler job.
type syncTask struct {
	*task
}

// newAsyncTask is a constructor func for synchronous job.
func newSyncTask(reqTimeout time.Duration) *syncTask {
	return &syncTask{task: &task{}}
}

// Do executes handler (handler should be a closure - otherwise you will not be
// able to pass arguments to it or call Complete() in order to return some value).
func (st *syncTask) Do(ctx context.Context, handler func(stop <-chan struct{}) error) {
	// memorize start time and change job status
	st.Lock()
	st.status, st.started = StatusInProgress, time.Now()
	st.Unlock()
	// error chan
	errChan := make(chan error, 1)
	// call handler in goroutine
	go func() { errChan <- handler(ctx.Done()) }()
	// wait until context deadline or job is done
	select {
	// job was done
	case err := <-errChan:
		st.Complete(nil, err)
	// or timeout is reached
	case <-ctx.Done():
		// ignore - 408 response will be sent by AsyncRequest func
	}
	return
}

// asyncTask represents asynchronous handler job.
type asyncTask struct {
	*task
	// unique request ID
	ID string
}

// newAsyncTask is a constructor func for asynchronous job.
func newAsyncTask(execTimeout time.Duration) *asyncTask {
	id := md5.Sum([]byte(time.Now().String()))
	return &asyncTask{
		ID: hex.EncodeToString(id[:]),
		task: &task{
			asyncTimeout: execTimeout,
		},
	}
}

// Do handles asynchronous execution of the handler.
func (at *asyncTask) Do(ctx context.Context, handler func(stop <-chan struct{}) error) {
	// memorize start time and change job status
	at.Lock()
	at.status, at.started = StatusInProgress, time.Now()
	at.Unlock()
	// error chan
	errChan := make(chan error, 1)
	// call handler in goroutine
	go func() {
		// call the handler with actual (execution) timeout channel
		errChan <- handler(func() <-chan struct{} {
			// context deadline channel
			ch := make(chan struct{}, 1)
			// run timer in a new goroutine
			go func() {
				<-time.NewTimer(at.asyncTimeout).C
				// channel may be closed after job is done (in some time)
				close(ch)
				// complete the task with context deadline error
				at.Complete(nil, context.DeadlineExceeded)
			}()
			return ch
		}())
	}()
	// wait until context deadline or job is done
	select {
	// job was done
	case err := <-errChan:
		// task should be completed in case if Complete has not been called in the
		// handler (for instance error was returned without wrapping with Complete)
		at.Complete(nil, err)
	// timeout
	case <-ctx.Done():
		// request timeout, this is not the error for async requests because handler
		// probably is still running
	}
	return
}

// AsyncRequest middleware provides a mechanism to request the data again after timeout.
// 	reqTimeout - time allotted for processing HTTP request, if request has not been
// processed completely - returns an ID of request (to retrieve result later).
// 	asyncTimeout - maximum time for async job to be done (actual context deadline),
// this logic should be implemented in asynchronous handler or skipped - in that case
// handler cannot be interrupted.
func AsyncRequest(reqTimeout, asyncTimeout, keepResult time.Duration) Middleware {
	// no sense to use this middleware if the following condition is not satisfied
	if !(reqTimeout < asyncTimeout && asyncTimeout < keepResult) {
		panic("request timeout should be less than async timeout and keep result should be greater than async timeout")
	}
	// each handler has its own task list
	var asyncJobs = timap.New(keepResult)
	// create a new Middleware
	return func(next http.Handler) http.Handler {
		// define the httprouter.Handle
		return ContextDeadline(reqTimeout)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// current job (can be sync/async)
			var currJob HandlerTask
			// check if async request, if not - ignore the next code block
			if _, ok := r.Header[asyncHeader]; ok {
				// current request
				var async *asyncTask
				// if contains ID - it is not a new request
				if requestID := r.Header.Get(asyncRequestID); requestID != "" {
					// find async job
					val, ok := asyncJobs.Load(requestID)
					if !ok {
						// async request is expired or has invalid ID
						http.Error(w, "invalid or expired request", http.StatusBadRequest)
						// skip next middleware/handlers
						return
					}
					async = val.(*asyncTask)
				} else {
					// create new async task
					async = newAsyncTask(asyncTimeout)
					//  and store in the list
					asyncJobs.Store(async.ID, async)
				}
				// set current job
				currJob = async
				// check async on exit and remove if it's done
				defer func() {
					if async.Status() == StatusDone {
						asyncJobs.Delete(async.ID)
					} else {
						// return request ID
						w.Header().Set(asyncRequestID, async.ID)
						// the status ot request is "accepted"
						w.WriteHeader(http.StatusAccepted)
						// provide a basic info message to the client
						w.Write([]byte("request is in progress\n"))
					}
				}()
			} else {
				// create synchronous job
				currJob = newSyncTask(reqTimeout)
				// send timeout code on exit if synchronous job was not done
				defer func() {
					if _, err := currJob.Resolve(); err == ErrNotCompleted {
						http.Error(w, context.DeadlineExceeded.Error(), http.StatusRequestTimeout)
					}
				}()
			}
			// get context from request
			ctx := r.Context()
			// put async task to the context
			ctx = context.WithValue(ctx, asyncContextKey, currJob)
			// replace request
			r = r.WithContext(ctx)
			// call next handler
			next.ServeHTTP(w, r)
		}))
	}
}

// GetHandlerTask extracts current job from context.
func GetHandlerTask(ctx context.Context) (HandlerTask, bool) {
	async, ok := ctx.Value(asyncContextKey).(HandlerTask)
	return async, ok
}
