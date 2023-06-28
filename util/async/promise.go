package async

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
