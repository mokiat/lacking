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
		scene:   scene,
		mask:    scene.newComponentType(),
		list:    mem.NewSparseList[T](1024),
		mapping: make([]mem.SparseID, scene.MaxEntityCount()),
	}
	scene.purgeSubscriptions.Subscribe(result.Unset)
	return result
}

var _ ComponentSet[any] = (*SparseComponentSet[any])(nil)

type SparseComponentSet[T any] struct {
	scene   *Scene
	mask    componentMask
	list    *mem.SparseList[T]
	mapping []mem.SparseID
}

func (s *SparseComponentSet[T]) Set(entityID EntityID, value T) {
	s.scene.assignComponent(entityID, s.mask)

	if id := s.mapping[entityID.index]; s.list.Has(id) {
		ref := s.list.Get(id)
		*ref = value
	} else {
		id, ref := s.list.New()
		*ref = value
		s.mapping[entityID.index] = id
	}
}

func (s *SparseComponentSet[T]) Unset(entityID EntityID) {
	s.scene.removeComponent(entityID, s.mask)

	id := s.mapping[entityID.index]
	s.list.Delete(id)
}

func (s *SparseComponentSet[T]) Ref(entityID EntityID) *T {
	if !s.scene.hasComponent(entityID, s.mask) {
		return nil
	}

	id := s.mapping[entityID.index]
	return s.list.Get(id)
}

func (s *SparseComponentSet[T]) Mask() componentMask {
	return s.mask
}
