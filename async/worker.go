package async

func NewWorker(capacity int) *Worker {
	return &Worker{
		tasks:   make(chan workerTask, capacity),
		flushed: make(chan struct{}),
	}
}

type Worker struct {
	tasks   chan workerTask
	flushed chan struct{}
}

func (w *Worker) Wait(task Task) Result {
	return w.Schedule(task).Wait()
}

func (w *Worker) Schedule(task Task) Outcome {
	outcome := NewOutcome()
	w.tasks <- workerTask{
		outcome: outcome,
		task:    task,
	}
	return outcome
}

func (w *Worker) ProcessTrySingle() bool {
	select {
	case task, ok := <-w.tasks:
		if !ok {
			return false
		}
		task.Run()
		return true
	default:
		return false
	}
}

func (w *Worker) ProcessTryMultiple(count int) bool {
	for count > 0 {
		if !w.ProcessTrySingle() {
			return false
		}
		count--
	}
	return true
}

func (w *Worker) ProcessAll() {
	for task := range w.tasks {
		task.Run()
	}
	close(w.flushed)
}

func (w *Worker) Shutdown() {
	close(w.tasks)
	w.ProcessAll()
	<-w.flushed
}

type workerTask struct {
	outcome Outcome
	task    Task
}

func (t workerTask) Run() {
	t.outcome.Record(t.task())
}
