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
		scene:      scene,
		mask:       scene.newComponentType(),
		components: make([]T, scene.MaxEntityCount()),
	}
}

var _ ComponentSet[any] = (*DenseComponentSet[any])(nil)

type DenseComponentSet[T any] struct {
	scene      *Scene
	mask       componentMask
	components []T
}

func (s *DenseComponentSet[T]) Set(entityID EntityID, value T) {
	s.scene.assignComponent(entityID, s.mask)

	s.components[entityID.index] = value
}

func (s *DenseComponentSet[T]) Unset(entityID EntityID) {
	s.scene.removeComponent(entityID, s.mask)

	s.components[entityID.index] = gog.Zero[T]()
}

func (s *DenseComponentSet[T]) Ref(entityID EntityID) *T {
	if !s.scene.hasComponent(entityID, s.mask) {
		return nil
	}

	return &s.components[entityID.index]
}

func (s *DenseComponentSet[T]) Mask() componentMask {
	return s.mask
}
