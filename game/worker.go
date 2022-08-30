package game

type Worker interface {
	Schedule(fn func() error) Operation
	ScheduleVoid(fn func()) Operation
}

type WorkerFunc func(fn func() error) Operation

func (f WorkerFunc) Schedule(fn func() error) Operation {
	return f(fn)
}

func (f WorkerFunc) ScheduleVoid(fn func()) Operation {
	return f(func() error {
		fn()
		return nil
	})
}

func NewOperation() Operation {
	return Operation{
		done: make(chan error),
	}
}

type Operation struct {
	done chan error
}

func (o Operation) Complete(err error) {
	o.done <- err
	close(o.done)
}

func (o Operation) Wait() error {
	return <-o.done
}
