package graphics

type Task struct {
	fn     func() error
	done   chan error
	errRun error
}

func (t *Task) Done() bool {
	select {
	case err, ok := <-t.done:
		if ok {
			t.errRun = err
		}
		return true
	default:
		return false
	}
}

func (t *Task) Wait() error {
	err := <-t.done
	t.errRun = err
	return err
}

func (t *Task) Error() error {
	return t.errRun
}

func NewWorker() *Worker {
	return &Worker{
		queue: make(chan *Task, 16),
	}
}

type Worker struct {
	queue chan *Task
}

func (w *Worker) Work() {
	select {
	case task := <-w.queue:
		err := task.fn()
		task.done <- err
		close(task.done)
	default:
	}
}

func (w *Worker) Schedule(fn func() error) *Task {
	task := &Task{
		fn:   fn,
		done: make(chan error, 1),
	}
	w.queue <- task
	return task
}

func (w *Worker) Run(fn func() error) error {
	task := w.Schedule(fn)
	task.Wait()
	return task.Error()
}
