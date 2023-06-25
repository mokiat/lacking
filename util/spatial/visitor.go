package spatial

// Visitor represents a callback mechanism to pass items back to the client.
type Visitor[T any] interface {
	// Visit is called for each observed item.
	Visit(item T)
}

// VisitorFunc is an implementation of Visitor that passes each observed
// item to the wrapped function.
type VisitorFunc[T any] func(item T)

// Visit calls the wrapped function.
func (f VisitorFunc[T]) Visit(item T) {
	f(item)
}

// NewVisitorBucket creates a new NewVisitorBucket instance with the specified
// initial capacity, which is only used to preallocate memory. It is allowed
// to exceed the initial capacity.
func NewVisitorBucket[T any](initCapacity int) *VisitorBucket[T] {
	return &VisitorBucket[T]{
		items: make([]T, 0, initCapacity),
	}
}

// VisitorBucket is an implementation of Visitor that stores observed items
// into a buffer for faster and more cache-friendly iteration afterwards.
type VisitorBucket[T any] struct {
	items []T
}

// Reset rewinds the item buffer.
func (r *VisitorBucket[T]) Reset() {
	r.items = r.items[:0]
}

// Visit records the passed item into the buffer.
func (r *VisitorBucket[T]) Visit(item T) {
	r.items = append(r.items, item)
}

// Each calls the provided closure function for each item in the buffer.
func (r *VisitorBucket[T]) Each(cb func(item T)) {
	for _, item := range r.items {
		cb(item)
	}
}

// Items returns the items stored in the buffer. The returned slice is valid
// only until the Reset function is called.
func (r *VisitorBucket[T]) Items() []T {
	return r.items
}
