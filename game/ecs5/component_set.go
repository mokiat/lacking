package ecs5

import (
	"github.com/mokiat/gog"
)

type ComponentSet[T any] interface {
	Set(entity Entity, value T)
	Unset(entity Entity)
	Ref(entity Entity) *T
	getMask() componentMask
}

func NewDenseComponentSet[T any](scene *Scene) *DenseComponentSet[T] {
	return &DenseComponentSet[T]{
		mask:       scene.newComponentType(),
		components: make([]T, scene.MaxEntityCount()),
	}
}

var _ ComponentSet[any] = (*DenseComponentSet[any])(nil)

type DenseComponentSet[T any] struct {
	mask       componentMask
	components []T
}

func (s *DenseComponentSet[T]) Set(entity Entity, value T) {
	scene := entity.scene

	handle := &scene.handles[entity.index]
	if handle.revision != entity.revision {
		panic("cannot add component to deleted entity")
	}
	handle.components |= s.mask

	s.components[entity.index] = value
}

func (s *DenseComponentSet[T]) Unset(entity Entity) {
	scene := entity.scene

	handle := &scene.handles[entity.index]
	if handle.revision != entity.revision {
		panic("cannot remove component from deleted entity")
	}
	handle.components &= ^s.mask

	s.components[entity.index] = gog.Zero[T]()
}

func (s *DenseComponentSet[T]) Ref(entity Entity) *T {
	scene := entity.scene

	handle := &scene.handles[entity.index]
	if handle.revision != entity.revision {
		panic("cannot reference component of deleted entity")
	}

	if (handle.components & s.mask) == 0 {
		return nil // entity does not have component
	}

	return &s.components[entity.index]
}

func (s *DenseComponentSet[T]) getMask() componentMask {
	return s.mask
}

func NewSparseComponentSet[T any](scene *Scene) ComponentSet[T] {
	panic("TODO")
}
