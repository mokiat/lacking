package async

import (
	"cmp"
	"errors"
)

func NewDeliveredPromise[T any](value T) Promise[T] {
	result := NewPromise[T]()
	result.Deliver(value)
	return result
}

func NewFailedPromise[T any](err error) Promise[T] {
	result := NewPromise[T]()
	result.Fail(err)
	return result
}

func NewPromise[T any]() Promise[T] {
	return Promise[T]{
		ch: make(chan promiseOutcome[T], 1),
	}
}

type Promise[T any] struct {
	ch chan promiseOutcome[T]
}

func (p Promise[T]) Ready() bool {
	select {
	case entry := <-p.ch:
		p.ch <- entry
		return true
	default:
		return false
	}
}

func (p Promise[T]) Wait() (T, error) {
	entry := <-p.ch
	p.ch <- entry
	return entry.value, entry.err
}

func (p Promise[T]) Inject(target *T) error {
	result, err := p.Wait()
	if err != nil {
		return err
	}
	*target = result
	return nil
}

func (p Promise[T]) Deliver(value T) {
	p.ch <- promiseOutcome[T]{
		value: value,
	}
}

func (p Promise[T]) Fail(err error) {
	p.ch <- promiseOutcome[T]{
		err: err,
	}
}

func (p Promise[T]) OnReady(cb func()) Promise[T] {
	go func() {
		p.Wait()
		cb()
	}()
	return p
}

func (p Promise[T]) OnSuccess(cb func(value T)) Promise[T] {
	go func() {
		if value, err := p.Wait(); err == nil {
			cb(value)
		}
	}()
	return p
}

func (p Promise[T]) OnError(cb func(err error)) Promise[T] {
	go func() {
		if _, err := p.Wait(); err != nil {
			cb(err)
		}
	}()
	return p
}

type promiseOutcome[T any] struct {
	value T
	err   error
}

func WaitPromises[T any](promises ...Promise[T]) ([]T, error) {
	var errs []error
	results := make([]T, len(promises))
	for i, promise := range promises {
		result, err := promise.Wait()
		if err != nil {
			errs = append(errs, err)
		} else {
			results[i] = result
		}
	}
	return results, errors.Join(errs...)
}

func NewFuncOperation(fn func() error) Operation {
	result := NewOperation()
	go func() {
		if err := fn(); err != nil {
			result.Fail(err)
		} else {
			result.Pass()
		}
	}()
	return result
}

func NewOperation() Operation {
	return Operation{
		promise: NewPromise[struct{}](),
	}
}

func NewPassedOperation() Operation {
	return Operation{
		NewDeliveredPromise(struct{}{}),
	}
}

func NewFailedOperation(err error) Operation {
	return Operation{
		promise: NewFailedPromise[struct{}](err),
	}
}

type Operation struct {
	promise Promise[struct{}]
}

func (o Operation) Pass() {
	o.promise.Deliver(struct{}{})
}

func (o Operation) Fail(err error) {
	o.promise.Fail(err)
}

func (o Operation) Wait() error {
	_, err := o.promise.Wait()
	return err
}

func (o Operation) OnSuccess(cb func()) Operation {
	go func() {
		if err := o.Wait(); err == nil {
			cb()
		}
	}()
	return o
}

func (o Operation) OnError(cb func(err error)) Operation {
	go func() {
		if err := o.Wait(); err != nil {
			cb(err)
		}
	}()
	return o
}

func (o Operation) IsCompleted() bool {
	return o.promise.Ready()
}

func InjectionPromise[T any](operation Operation, target T) Promise[T] {
	result := NewPromise[T]()
	go func() {
		if err := operation.Wait(); err == nil {
			result.Deliver(target)
		} else {
			result.Fail(err)
		}
	}()
	return result
}

func WaitOperations(operations ...Operation) error {
	var err error
	for _, operation := range operations {
		err = cmp.Or(err, operation.Wait())
	}
	return err
}

func JoinOperations(operations ...Operation) Operation {
	return NewFuncOperation(func() error {
		var err error
		for _, operation := range operations {
			err = cmp.Or(err, operation.Wait())
		}
		return err
	})
}

func Sequential(actions ...func() Operation) Operation {
	return NewFuncOperation(func() error {
		for _, action := range actions {
			if err := action().Wait(); err != nil {
				return err
			}
		}
		return nil
	})
}
