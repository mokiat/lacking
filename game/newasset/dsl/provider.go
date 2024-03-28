package dsl

import (
	"sync"
)

type Provider[T any] interface {
	Get() (T, error)
	Digest() ([]byte, error)
}

func OnceProvider[T any](delegate Provider[T]) Provider[T] {
	return &onceProvider[T]{
		getFunc: sync.OnceValues(func() (T, error) {
			return delegate.Get()
		}),
		digestFunc: sync.OnceValues(func() ([]byte, error) {
			return delegate.Digest()
		}),
	}
}

type onceProvider[T any] struct {
	getFunc    func() (T, error)
	digestFunc func() ([]byte, error)
}

func (p *onceProvider[T]) Get() (T, error) {
	return p.getFunc()
}

func (p *onceProvider[T]) Digest() ([]byte, error) {
	return p.digestFunc()
}
