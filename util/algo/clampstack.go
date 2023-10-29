package algo

func NewClampStack[T any](initCapacity int) *ClampStack[T] {
	return &ClampStack[T]{
		items: make([]T, 0, initCapacity),
	}
}

type ClampStack[T any] struct {
	items []T
}

func (s *ClampStack[T]) Size() int {
	return len(s.items)
}

func (s *ClampStack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *ClampStack[T]) Push(v T) {
	if len(s.items) == cap(s.items) {
		copy(s.items[0:], s.items[1:]) // shift left
		s.items[len(s.items)-1] = v
	} else {
		s.items = append(s.items, v)
	}
}

func (s *ClampStack[T]) Pop() T {
	count := len(s.items)
	result := s.items[count-1]
	s.items = s.items[:count-1]
	return result
}

func (s *ClampStack[T]) Peek() T {
	return s.items[len(s.items)-1]
}

func (s *ClampStack[T]) Clear() {
	s.items = s.items[:0]
}
