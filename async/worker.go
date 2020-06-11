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

func (w *Worker) Schedule(task Task) Outcome {
	outcome := NewOutcome()
	w.tasks <- workerTask{
		outcome: outcome,
		task:    task,
	}
	return outcome
}

func (w *Worker) Work() {
	for task := range w.tasks {
		task.Run()
	}
	close(w.flushed)
}

func (w *Worker) Shutdown() {
	close(w.tasks)
	<-w.flushed
}

type workerTask struct {
	outcome Outcome
	task    Task
}

func (t workerTask) Run() {
	t.outcome.Record(t.task())
}
