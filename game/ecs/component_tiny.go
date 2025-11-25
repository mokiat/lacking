package ecs

import (
	"github.com/mokiat/lacking/util/mem"
)

// NewTinyComponentSet returns a ComponentSet implementation that allocates
// very little storage for components.
//
// This implementation is the most memory-friendly but this comes at a huge
// performance cost. It should be used only for component types that are rerely
// ever attached to an entity.
func NewTinyComponentSet[T any](scene *Scene) *TinyComponentSet[T] {
	result := &TinyComponentSet[T]{
		scene:   scene,
		mask:    scene.newComponentType(),
		list:    mem.NewSparseList[T](16),
		mapping: make(map[uint32]mem.SparseID),
	}
	scene.purgeSubscriptions.Subscribe(result.Unset)
	return result
}

var _ ComponentSet[any] = (*TinyComponentSet[any])(nil)

type TinyComponentSet[T any] struct {
	scene   *Scene
	mask    componentMask
	list    *mem.SparseList[T]
	mapping map[uint32]mem.SparseID
}

func (s *TinyComponentSet[T]) Set(entityID EntityID, value T) {
	s.Unset(entityID)

	s.scene.assignComponent(entityID, s.mask)

	if id, ok := s.mapping[entityID.index]; ok {
		ref := s.list.Get(id)
		*ref = value
	} else {
		id, ref := s.list.New()
		*ref = value
		s.mapping[entityID.index] = id
	}
}

func (s *TinyComponentSet[T]) Unset(entityID EntityID) {
	s.scene.removeComponent(entityID, s.mask)

	if id, ok := s.mapping[entityID.index]; ok {
		s.list.Delete(id)
		delete(s.mapping, entityID.index)
	}
}

func (s *TinyComponentSet[T]) Ref(entityID EntityID) *T {
	if !s.scene.hasComponent(entityID, s.mask) {
		return nil
	}

	id, ok := s.mapping[entityID.index]
	if !ok {
		return nil
	}
	return s.list.Get(id)
}

func (s *TinyComponentSet[T]) Mask() componentMask {
	return s.mask
}
