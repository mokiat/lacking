package game

import "github.com/mokiat/lacking/util/async"

func newAsyncOperation(asyncEngine *AsyncEngine, delegate async.Operation) AsyncOperation {
	return AsyncOperation{
		asyncEngine: asyncEngine,
		delegate:    delegate,
	}
}

// AsyncOperation represents an asynchronous operation that is running on a
// worker thread. However, the methods on this object are safe to call from the
// main thread.
type AsyncOperation struct {
	asyncEngine *AsyncEngine
	delegate    async.Operation
}

// OnSuccess registers a callback that will be called when the operation
// completes successfully. The callback will be called on the main thread.
func (o AsyncOperation) OnSuccess(cb func()) {
	o.delegate.OnSuccess(func() {
		o.asyncEngine.ScheduleMain(func(*Engine) error {
			cb()
			return nil
		})
	})
}

// OnError registers a callback that will be called when the operation
// fails. The callback will be called on the main thread.
func (o AsyncOperation) OnError(cb func(error)) {
	o.delegate.OnError(func(err error) {
		o.asyncEngine.ScheduleMain(func(*Engine) error {
			cb(err)
			return nil
		})
	})
}

// OnCompleted registers a callback that will be called when the operation
// completes. The callback will be called on the main thread.
func (e AsyncOperation) OnCompleted(cb func(error)) {
	e.delegate.OnSuccess(func() {
		e.asyncEngine.ScheduleMain(func(*Engine) error {
			cb(nil)
			return nil
		})
	})
	e.delegate.OnError(func(err error) {
		e.asyncEngine.ScheduleMain(func(*Engine) error {
			cb(err)
			return nil
		})
	})
}

// IsCompleted returns true if the operation has completed, either
// successfully or with an error.
func (o AsyncOperation) IsCompleted() bool {
	return o.delegate.IsCompleted()
}
