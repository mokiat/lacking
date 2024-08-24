package async

import (
	"sync"
	"sync/atomic"
	"time"
)

type Func = func()

func NewWorker(capacity int) *Worker {
	return &Worker{
		tasks:   make(chan Func, capacity),
		running: 1,
	}
}

type Worker struct {
	tasks     chan Func
	running   int32
	pipelines sync.WaitGroup
}

func (w *Worker) Schedule(fn Func) {
	w.tasks <- fn
}

func (w *Worker) ProcessCount(count int) bool {
	w.pipelines.Add(1)
	defer w.pipelines.Done()
	if atomic.LoadInt32(&w.running) == 0 {
		return false
	}

	for count > 0 {
		if !w.processNextTask() {
			return false
		}
		count--
	}
	return true
}

func (w *Worker) ProcessDuration(targetDuration time.Duration) bool {
	w.pipelines.Add(1)
	defer w.pipelines.Done()
	if atomic.LoadInt32(&w.running) == 0 {
		return false
	}

	startTime := time.Now()
	for time.Since(startTime) < targetDuration {
		if !w.processNextTask() {
			return false
		}
	}
	return true
}

func (w *Worker) ProcessAll() {
	w.pipelines.Add(1)
	defer w.pipelines.Done()
	if atomic.LoadInt32(&w.running) == 0 {
		return
	}

	for task := range w.tasks {
		task()
	}
}

func (w *Worker) Shutdown() {
	atomic.StoreInt32(&w.running, 0)
	close(w.tasks)
	w.pipelines.Wait()
	for task := range w.tasks {
		task()
	}
}

func (w *Worker) processNextTask() bool {
	select {
	case task, ok := <-w.tasks:
		if ok {
			task()
			return true
		}
		return false
	default:
		return false
	}
}
