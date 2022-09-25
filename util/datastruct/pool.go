package datastruct

// Pool represents a storage structure that can preseve allocated objects
// for faster reuse.
type Pool[T any] interface {
	// Fetch retrieves an available item from the pool or creates a new one
	// if one is not available.
	Fetch() *T

	// Restore returns an item to the pool.
	Restore(*T)
}

// NewDynamicPool creates a new DynamicPool instance.
func NewDynamicPool[T any]() *DynamicPool[T] {
	return &DynamicPool[T]{}
}

var _ Pool[any] = (*DynamicPool[any])(nil)

// DynamicPool is an implementation of Pool that caches restored
// items into a list. Fetching an item tries to find an existing
// reference from the list, otherwise allocates a new one.
type DynamicPool[T any] struct {
	items []*T
}

func (p *DynamicPool[T]) Fetch() *T {
	if count := len(p.items); count > 0 {
		result := p.items[count-1]
		p.items = p.items[:count-1]
		return result
	}
	return new(T)
}

func (p *DynamicPool[T]) Restore(v *T) {
	p.items = append(p.items, v)
}

// NewStaticPool creates a new StaticPool instance with the specified
// capacity.
func NewStaticPool[T any](capacity int) *StaticPool[T] {
	result := &StaticPool[T]{
		items: make([]*T, capacity),
	}
	flat := make([]T, capacity)
	for i := 0; i < len(flat); i++ {
		result.items[i] = &flat[i]
	}
	return result
}

var _ Pool[any] = (*StaticPool[any])(nil)

// StaticPool is an implementation of Pool that tries to allocate items
// next to each other for improved cache locality.
type StaticPool[T any] struct {
	items []*T
}

func (p *StaticPool[T]) Fetch() *T {
	count := len(p.items)
	result := p.items[count-1]
	p.items = p.items[:count-1]
	return result
}

func (p *StaticPool[T]) Restore(v *T) {
	p.items = append(p.items, v)
}
