package datastruct

func NewStack[T any](initCapacity int) *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0, initCapacity),
	}
}

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() T {
	count := len(s.items)
	if count == 0 {
		panic("stack is empty")
	}
	result := s.items[count-1]
	s.items = s.items[:count-1]
	return result
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}
