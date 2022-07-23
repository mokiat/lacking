package datastruct

// NewStack creates a new Stack instance with the specified initial capacity,
// which only serves to preallocate memory. Exceeding the initial capacity is
// allowed.
func NewStack[T any](initCapacity int) *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0, initCapacity),
	}
}

// Stack is an implementation of a stack datastructure. The last inserted
// item is the first one to be removed (LIFO - last in, first out).
type Stack[T any] struct {
	items []T
}

// Push adds an item to this Stack.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes the last item to be added from this Stack.
// This function panics if there are no more items. Use IsEmpty to check
// for that.
func (s *Stack[T]) Pop() T {
	count := len(s.items)
	result := s.items[count-1]
	s.items = s.items[:count-1]
	return result
}

// IsEmpty returns true if there are no more items in this Stack.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}
