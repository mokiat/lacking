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

// NewVisitorBucket creates a new VisitorBucket instance with the specified
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

// PairVisitor represents a callback mechanism to pass pairs of items back to
// the client.
type PairVisitor[T any] interface {
	// Visit is called for each observed pair of items.
	Visit(first, second T)
}

// PairVisitorFunc is an implementation of PairVisitor that passes each observed
// pair of items to the wrapped function.
type PairVisitorFunc[T any] func(first, second T)

// Visit calls the wrapped function.
func (f PairVisitorFunc[T]) Visit(first, second T) {
	f(first, second)
}

// NewPairVisitorBucket creates a new PairVisitorBucket instance with the
// specified initial capacity, which is only used to preallocate memory.
// It is allowed to exceed the initial capacity.
func NewPairVisitorBucket[T any](initCapacity int) *PairVisitorBucket[T] {
	return &PairVisitorBucket[T]{
		items: make([]pairVisitorItem[T], 0, initCapacity),
	}
}

// PairVisitorBucket is an implementation of PairVisitor that stores observed
// item pairs into a buffer for faster and more cache-friendly iteration
// afterwards.
type PairVisitorBucket[T any] struct {
	items []pairVisitorItem[T]
}

// Reset rewinds the item buffer.
func (r *PairVisitorBucket[T]) Reset() {
	r.items = r.items[:0]
}

// Visit records the passed item pair into the buffer.
func (r *PairVisitorBucket[T]) Visit(first, second T) {
	r.items = append(r.items, pairVisitorItem[T]{
		first:  first,
		second: second,
	})
}

// Each calls the provided closure function for each item pair in the buffer.
func (r *PairVisitorBucket[T]) Each(cb func(first, second T)) {
	for _, item := range r.items {
		cb(item.first, item.second)
	}
}

type pairVisitorItem[T any] struct {
	first  T
	second T
}
