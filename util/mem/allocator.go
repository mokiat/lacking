package mem

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
)

// Allocator represents a mechanism to allocate and reuse objects in Go in a
// more predictable manner, similar to a pool.
type Allocator[T any] interface {

	// Allocate returns a new reference of T. This reference may be a brand
	// new one or it could be a reused one from before.
	Allocate() *T

	// Release returns the reference to the allocator. It may then be
	// returned by a future call to Allocate. Once Release is called the
	// reference should no longer be used by the caller until returned
	// back from Allocate.
	Release(*T)
}

var _ Allocator[any] = (*SparseAllocator[any])(nil)

// SparseAllocator is an Allocator implementation that acts as a pool.
// There is no guarantee of locality between allocated objects.
type SparseAllocator[T any] struct {
	pool *ds.Stack[*T]
}

// NewSparseAllocator returns a new SparseAllocator instance.
func NewSparseAllocator[T any]() *SparseAllocator[T] {
	return &SparseAllocator[T]{
		pool: ds.NewStack[*T](0),
	}
}

func (a *SparseAllocator[T]) Allocate() *T {
	if a.pool.IsEmpty() {
		return gog.PtrOf(gog.Zero[T]())
	}
	return a.pool.Pop()
}

func (a *SparseAllocator[T]) Release(ref *T) {
	a.pool.Push(ref)
}

var _ Allocator[any] = (*StaticAllocator[any])(nil)

// StaticAllocator is an Allocator implementation that manages a fixed number
// of closely allocated objects. Trying to allocate above the initial size
// will cause a panic. On the plus size, objects should be close in memory.
type StaticAllocator[T any] struct {
	pool *ds.Stack[*T]
}

// NewStaticAllocator returns a new StaticAllocator instance with the
// specified size.
func NewStaticAllocator[T any](size int) *StaticAllocator[T] {
	pool := ds.NewStack[*T](size)
	items := make([]T, size)
	for i := range items {
		pool.Push(&items[i])
	}
	return &StaticAllocator[T]{
		pool: pool,
	}
}

func (a *StaticAllocator[T]) Allocate() *T {
	if a.pool.IsEmpty() {
		panic("allocator has been exhausted")
	}
	return a.pool.Pop()
}

func (a *StaticAllocator[T]) Release(ref *T) {
	a.pool.Push(ref)
}

var _ Allocator[any] = (*BatchAllocator[any])(nil)

// BatchAllocator is an Allocator implementation that preallocates objects
// in batches of a given size. This allows objects returned from Allocate
// to be fairly close in memory while also not being restrictive on the size.
type BatchAllocator[T any] struct {
	pool      *ds.Stack[*T]
	batchSize int
}

// NewBatchAllocator returns a new BatchAllocator instance.
func NewBatchAllocator[T any](batchSize int) *BatchAllocator[T] {
	return &BatchAllocator[T]{
		pool:      ds.NewStack[*T](0),
		batchSize: batchSize,
	}
}

func (a *BatchAllocator[T]) Allocate() *T {
	if a.pool.IsEmpty() {
		batch := make([]T, a.batchSize)
		for i := range batch {
			a.pool.Push(&batch[i])
		}
	}
	return a.pool.Pop()
}

func (a *BatchAllocator[T]) Release(ref *T) {
	a.pool.Push(ref)
}
