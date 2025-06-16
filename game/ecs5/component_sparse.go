package ecs5

import (
	"github.com/mokiat/lacking/util/mem"
)

// NewSparseComponentSet returns a ComponentSet implementation that allocates
// storage as needed.
//
// This implementation is more memory-friendly but this comes at a performance
// cost. It should be used only for component types that are rarely attached
// to entities.
func NewSparseComponentSet[T any](scene *Scene) *SparseComponentSet[T] {
	result := &SparseComponentSet[T]{
		mask:    scene.newComponentType(),
		list:    mem.NewSparseList[T](64),
		mapping: make(map[uint32]mem.SparseID),
	}
	scene.SubscribeDelete(result.Unset)
	return result
}

var _ ComponentSet[any] = (*SparseComponentSet[any])(nil)

type SparseComponentSet[T any] struct {
	mask    componentMask
	list    *mem.SparseList[T]
	mapping map[uint32]mem.SparseID
}

func (s *SparseComponentSet[T]) Set(entity Entity, value T) {
	scene := entity.scene
	scene.assignComponent(entity, s.mask)

	id, ref := s.list.New()
	*ref = value
	s.mapping[entity.index] = id
}

func (s *SparseComponentSet[T]) Unset(entity Entity) {
	scene := entity.scene
	scene.removeComponent(entity, s.mask)

	if id, ok := s.mapping[entity.index]; ok {
		s.list.Delete(id)
	}
}

func (s *SparseComponentSet[T]) Ref(entity Entity) *T {
	scene := entity.scene
	if !scene.hasComponent(entity, s.mask) {
		return nil
	}

	id, ok := s.mapping[entity.index]
	if !ok {
		return nil
	}
	return s.list.Get(id)
}

func (s *SparseComponentSet[T]) Mask() componentMask {
	return s.mask
}
