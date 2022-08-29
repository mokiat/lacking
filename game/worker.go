package game

type Worker interface {
	Schedule(fn func())
}

type WorkerFunc func(fn func())

func (f WorkerFunc) Schedule(fn func()) {
	f(fn)
}
