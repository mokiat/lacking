package dsl

import "sync"

// GetFunc is a function that returns a value of a specific type.
type GetFunc[T any] func() (T, error)

// Provider is an interface that provides a value of a specific type.
type Provider[T any] interface {
	Digestable
	Get() (T, error)
}

// FuncProvider creates a provider from a get and a digest function.
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

// OnceProvider creates a provider that caches the result of the delegate
// provider.
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
