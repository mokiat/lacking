package shape2d

// VisitorFunc is a mechanism to receive items from a query.
type VisitorFunc[T any] func(item T) bool

// NewVisitorBucket creates a new VisitorBucket instance with the specified
// initial capacity, which is only used to preallocate memory (it is allowed to
// exceed the initial capacity).
func NewVisitorBucket[T any](initCapacity int) *VisitorBucket[T] {
	return &VisitorBucket[T]{
		items: make([]T, 0, initCapacity),
	}
}

// VisitorBucket can be used to store items returned from a query.
type VisitorBucket[T any] struct {
	items []T
}

// Reset clears any stored items.
func (r *VisitorBucket[T]) Reset() {
	clear(r.items)
	r.items = r.items[:0]
}

// VisitorFunc returns a VisitorFunc that can be passed to query functions.
//
// Make sure to call Reset before reusing the bucket, unless you want to
// append to the existing items.
func (r *VisitorBucket[T]) VisitorFunc() VisitorFunc[T] {
	return r.Add
}

// Add records the passed item into the bucket.
func (r *VisitorBucket[T]) Add(item T) bool {
	r.items = append(r.items, item)
	return true
}

// Each calls the provided closure function for each stored item.
func (r *VisitorBucket[T]) Each(yield func(item T) bool) {
	for _, item := range r.items {
		if !yield(item) {
			return
		}
	}
}

// Items returns the underlying slice of stored items. The returned slice is
// valid only until the Reset function is called.
func (r *VisitorBucket[T]) Items() []T {
	return r.items
}
