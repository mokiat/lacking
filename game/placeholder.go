package game

import "github.com/mokiat/lacking/async"

func pendingPlaceholder[T any]() Placeholder[T] {
	return Placeholder[T]{
		promise: async.NewPromise[T](),
	}
}

func failedPlaceholder[T any](err error) Placeholder[T] {
	return Placeholder[T]{
		promise: async.NewFailedPromise[T](err),
	}
}

type Placeholder[T any] struct {
	promise async.Promise[T]
}

func (p Placeholder[T]) Get() (T, error) {
	if !p.promise.Ready() {
		var zeroT T
		return zeroT, ErrStillLoading
	}
	return p.promise.Wait()
}

func (p Placeholder[T]) OnFinished(fn func(value T, err error)) {
	p.promise.OnSuccess(func(value T) {
		fn(value, nil)
	})
	p.promise.OnError(func(err error) {
		var zeroT T
		fn(zeroT, err)
	})
}
