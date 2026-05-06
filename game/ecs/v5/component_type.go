package ecs

import (
	"github.com/mokiat/lacking/game/ecs/v5/internal"
)

// BaseComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type BaseComponentType interface {

	// BaseStorage returns the component storage associated with the component
	// type.
	BaseStorage() internal.BaseComponentStorage
}

// ComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type ComponentType[T any] struct {
	id      internal.TypeID
	storage *internal.ComponentStorage[T]
}

var _ BaseComponentType = (*ComponentType[any])(nil)

// BaseStorage returns the component storage associated with the component
// type.
func (t ComponentType[T]) BaseStorage() internal.BaseComponentStorage {
	return t.storage
}
