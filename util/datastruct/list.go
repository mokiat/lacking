package datastruct

import "golang.org/x/exp/slices"

func NewList[T comparable]() *List[T] {
	return &List[T]{}
}

type List[T comparable] struct {
	items []T
}

func (l *List[T]) Size() int {
	return len(l.items)
}

func (l *List[T]) Add(item T) {
	l.items = append(l.items, item)
}

func (l *List[T]) Remove(item T) bool {
	index := l.IndexOf(item)
	if index < 0 {
		return false
	}
	l.items = slices.Delete(l.items, index, index+1)
	return true
}

func (l *List[T]) Get(index int) T {
	return l.items[index]
}

func (l *List[T]) Items() []T {
	return l.items
}

func (l *List[T]) Contains(item T) bool {
	return l.IndexOf(item) >= 0
}

func (l *List[T]) IndexOf(item T) int {
	return slices.Index(l.items, item)
}

func (l *List[T]) Each(iterator func(item T)) {
	for _, item := range l.items {
		iterator(item)
	}
}

func (l *List[T]) Clear() {
	l.items = l.items[:0]
}

func (l *List[T]) Clip() {
	l.items = slices.Clip(l.items)
}
