package resource

type Worker interface {
	Schedule(fn func() error)
}

func newTask(fn func() error) Task {
	return Task{
		fn:   fn,
		done: make(chan error, 1),
	}
}

type Task struct {
	fn   func() error
	done chan error
}

func (t Task) Run() error {
	err := t.fn()
	t.done <- err
	return err
}

func (t Task) Wait() error {
	return <-t.done
}
