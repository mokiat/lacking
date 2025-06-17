package ecs

import (
	"github.com/mokiat/gog"
)

// NewDenseComponentSet returns a ComponentSet implementation that has
// pre-allocated storage for the maximum number of entities.
//
// While this implementation is the fastest available, it is also the most
// memory intensive and should be used only for components that are very
// common and are likely to be attached to the majority of entities.
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
	scene.assignComponent(entity, s.mask)

	s.components[entity.index] = value
}

func (s *DenseComponentSet[T]) Unset(entity Entity) {
	scene := entity.scene
	scene.removeComponent(entity, s.mask)

	s.components[entity.index] = gog.Zero[T]()
}

func (s *DenseComponentSet[T]) Ref(entity Entity) *T {
	scene := entity.scene
	if !scene.hasComponent(entity, s.mask) {
		return nil
	}

	return &s.components[entity.index]
}

func (s *DenseComponentSet[T]) Mask() componentMask {
	return s.mask
}
