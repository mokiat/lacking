package ecs5

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
		mask:    scene.newComponentType(),
		list:    mem.NewSparseList[T](16),
		mapping: make(map[uint32]mem.SparseID),
	}
	scene.SubscribeDelete(result.Unset)
	return result
}

var _ ComponentSet[any] = (*TinyComponentSet[any])(nil)

type TinyComponentSet[T any] struct {
	mask    componentMask
	list    *mem.SparseList[T]
	mapping map[uint32]mem.SparseID
}

func (s *TinyComponentSet[T]) Set(entity Entity, value T) {
	s.Unset(entity)

	scene := entity.scene
	scene.assignComponent(entity, s.mask)

	if id, ok := s.mapping[entity.index]; ok {
		ref := s.list.Get(id)
		*ref = value
	} else {
		id, ref := s.list.New()
		*ref = value
		s.mapping[entity.index] = id
	}
}

func (s *TinyComponentSet[T]) Unset(entity Entity) {
	scene := entity.scene
	scene.removeComponent(entity, s.mask)

	if id, ok := s.mapping[entity.index]; ok {
		s.list.Delete(id)
		delete(s.mapping, entity.index)
	}
}

func (s *TinyComponentSet[T]) Ref(entity Entity) *T {
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

func (s *TinyComponentSet[T]) Mask() componentMask {
	return s.mask
}
