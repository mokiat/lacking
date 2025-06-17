package ecs

import (
	"github.com/mokiat/lacking/util/mem"
)

// NewSparseComponentSet returns a ComponentSet implementation that allocates
// storage as needed.
//
// This implementation is more memory-friendly but this comes at a performance
// cost. It should be used only for component types that are occasionally
// attached to an entity.
func NewSparseComponentSet[T any](scene *Scene) *SparseComponentSet[T] {
	result := &SparseComponentSet[T]{
		mask:    scene.newComponentType(),
		list:    mem.NewSparseList[T](1024),
		mapping: make([]mem.SparseID, scene.MaxEntityCount()),
	}
	scene.SubscribeDelete(result.Unset)
	return result
}

var _ ComponentSet[any] = (*SparseComponentSet[any])(nil)

type SparseComponentSet[T any] struct {
	mask    componentMask
	list    *mem.SparseList[T]
	mapping []mem.SparseID
}

func (s *SparseComponentSet[T]) Set(entity Entity, value T) {
	scene := entity.scene
	scene.assignComponent(entity, s.mask)

	if id := s.mapping[entity.index]; s.list.Has(id) {
		ref := s.list.Get(id)
		*ref = value
	} else {
		id, ref := s.list.New()
		*ref = value
		s.mapping[entity.index] = id
	}
}

func (s *SparseComponentSet[T]) Unset(entity Entity) {
	scene := entity.scene
	scene.removeComponent(entity, s.mask)

	id := s.mapping[entity.index]
	s.list.Delete(id)
}

func (s *SparseComponentSet[T]) Ref(entity Entity) *T {
	scene := entity.scene
	if !scene.hasComponent(entity, s.mask) {
		return nil
	}

	id := s.mapping[entity.index]
	return s.list.Get(id)
}

func (s *SparseComponentSet[T]) Mask() componentMask {
	return s.mask
}
