package async

import (
	"sync"
	"sync/atomic"
	"time"
)

type Func func() error

func NewWorker(capacity int) *Worker {
	return &Worker{
		tasks:   make(chan workerTask, capacity),
		running: 1,
	}
}

type Worker struct {
	tasks     chan workerTask
	running   int32
	pipelines sync.WaitGroup
}

func (w *Worker) Schedule(fn Func) Promise[error] {
	promise := NewPromise[error]()
	w.tasks <- workerTask{
		fn:      fn,
		promise: promise,
	}
	return promise
}

func (w *Worker) ProcessCount(count int) bool {
	if atomic.LoadInt32(&w.running) == 0 {
		return false
	}
	w.pipelines.Add(1)
	defer w.pipelines.Done()

	for count > 0 {
		if !w.processNextTask() {
			return false
		}
		count--
	}
	return true
}

func (w *Worker) ProcessDuration(targetDuration time.Duration) bool {
	if atomic.LoadInt32(&w.running) == 0 {
		return false
	}
	w.pipelines.Add(1)
	defer w.pipelines.Done()

	startTime := time.Now()
	for time.Since(startTime) < targetDuration {
		if !w.processNextTask() {
			return false
		}
	}
	return true
}

func (w *Worker) ProcessAll() {
	if atomic.LoadInt32(&w.running) == 0 {
		return
	}
	w.pipelines.Add(1)
	defer w.pipelines.Done()

	for task := range w.tasks {
		task.Run()
	}
}

func (w *Worker) Shutdown() {
	atomic.StoreInt32(&w.running, 0)
	close(w.tasks)
	for task := range w.tasks {
		task.Run()
	}
	w.pipelines.Wait()
}

func (w *Worker) processNextTask() bool {
	select {
	case task, ok := <-w.tasks:
		if ok {
			task.Run()
			return true
		}
		return false
	default:
		return false
	}
}

type workerTask struct {
	fn      Func
	promise Promise[error]
}

func (t workerTask) Run() {
	t.promise.Deliver(t.fn())
}
