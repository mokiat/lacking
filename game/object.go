package game

import (
	"fmt"
	"iter"

	"github.com/mokiat/gog"
)

const UnspecifiedID = uint32(0xFFFFFFFF)

// Identifiable is a generic type that holds an ID and a value of type T.
//
// It is used to represent objects loaded from an asset file. The ID is not
// globally unique and should be used in the scope of other objects from
// the same asset file.
type Identifiable[T any] struct {
	ID    uint32
	Value T
}

type IdentifiableList[T any] []Identifiable[T]

func (l IdentifiableList[T]) Iter() iter.Seq2[uint32, T] {
	return func(yield func(uint32, T) bool) {
		for _, entry := range l {
			if !yield(entry.ID, entry.Value) {
				break
			}
		}
	}
}

func (l IdentifiableList[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, entry := range l {
			if !yield(entry.Value) {
				break
			}
		}
	}
}

func (l IdentifiableList[T]) ValuesList() []T {
	result := make([]T, len(l))
	for i, entry := range l {
		result[i] = entry.Value
	}
	return result
}

func (l IdentifiableList[T]) FindByID(id uint32) (T, bool) {
	for _, item := range l {
		if item.ID == id {
			return item.Value, true
		}
	}
	return gog.Zero[T](), false
}

func (l IdentifiableList[T]) GetByID(id uint32) T {
	for _, item := range l {
		if item.ID == id {
			return item.Value
		}
	}
	panic(fmt.Errorf("item with ID %d not found", id))
}
