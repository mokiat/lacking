package game

type Worker interface {
	Schedule(fn func()) Operation
}

type WorkerFunc func(fn func()) Operation

func (f WorkerFunc) Schedule(fn func()) Operation {
	return f(fn)
}

func NewOperation() Operation {
	return Operation{
		done: make(chan struct{}),
	}
}

type Operation struct {
	done chan struct{}
}

func (o Operation) Complete() {
	close(o.done)
}

func (o Operation) Wait() {
	<-o.done
}
