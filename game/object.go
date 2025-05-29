package game

import (
	"fmt"

	"github.com/mokiat/gog"
)

const UnspecifiedID = uint32(0xFFFFFFFF)

type Identifiable[T any] struct {
	ID    uint32
	Value T
}

type IdentifiableList[T any] []Identifiable[T]

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
