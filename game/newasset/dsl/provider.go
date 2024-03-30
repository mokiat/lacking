package dsl

import "sync"

type GetFunc[T any] func() (T, error)

type Provider[T any] interface {
	Get() (T, error)
	Digest() ([]byte, error)
}

func FuncProvider[T any](get GetFunc[T], digest DigestFunc) Provider[T] {
	return &funcProvider[T]{
		getFunc:    get,
		digestFunc: digest,
	}
}

type funcProvider[T any] struct {
	getFunc    GetFunc[T]
	digestFunc DigestFunc
}

func (p *funcProvider[T]) Get() (T, error) {
	return p.getFunc()
}

func (p *funcProvider[T]) Digest() ([]byte, error) {
	return p.digestFunc()
}

func OnceProvider[T any](delegate Provider[T]) Provider[T] {
	return FuncProvider[T](
		sync.OnceValues(func() (T, error) {
			return delegate.Get()
		}),
		sync.OnceValues(func() ([]byte, error) {
			return delegate.Digest()
		}),
	)
}
