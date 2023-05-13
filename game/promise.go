package game

import (
	"errors"

	"github.com/mokiat/lacking/util/async"
)

var ErrNotReady = errors.New("not ready")

func NewPromise[T any](engine *Engine) Promise[T] {
	return SafePromise(async.NewPromise[T](), engine)
}

func SafePromise[T any](delegate async.Promise[T], engine *Engine) Promise[T] {
	return Promise[T]{
		delegate: delegate,
		worker:   engine.GFXWorker(),
	}
}

type Promise[T any] struct {
	delegate async.Promise[T]
	worker   Worker
}

func (p Promise[T]) Unsafe() async.Promise[T] {
	return p.delegate
}

func (p Promise[T]) Ready() bool {
	return p.delegate.Ready()
}

func (p Promise[T]) OnReady(cb func()) {
	go func() {
		p.delegate.Wait()
		p.worker.ScheduleVoid(func() {
			cb()
		})
	}()
}

func (p Promise[T]) Get() (T, error) {
	if !p.Ready() {
		var v T
		return v, ErrNotReady
	}
	return p.delegate.Wait()
}

func (p Promise[T]) OnSuccess(cb func(value T)) {
	go func() {
		if value, err := p.delegate.Wait(); err == nil {
			p.worker.ScheduleVoid(func() {
				cb(value)
			})
		}
	}()
}

func (p Promise[T]) OnError(cb func(err error)) {
	go func() {
		if _, err := p.delegate.Wait(); err != nil {
			p.worker.ScheduleVoid(func() {
				cb(err)
			})
		}
	}()
}
