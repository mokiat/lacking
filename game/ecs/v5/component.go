package ecs

import (
	"reflect"

	"github.com/mokiat/lacking/game/ecs/v5/internal"
)

// NewScope creates a new scope for component type registration.
func NewScope() *Scope {
	return &Scope{
		registeredTypes: make(map[reflect.Type]BaseComponentType),
	}
}

// Scope represents a scope for component type registration.
type Scope struct {
	registeredTypes map[reflect.Type]BaseComponentType
	componentTypes  [internal.MaxComponentTypes]BaseComponentType
}

// Type register the specified Go structure as a component type within
// the specified scope and returns its unique identifier. The identifier is used
// to refer to the component type in various API calls. It also acts as a
// storage for the component's data.
//
// This function should be called from a global variable initializer, and
// is not safe for concurrent use.
func Type[T any](scope *Scope) ComponentType[T] {
	if len(scope.registeredTypes) >= internal.MaxComponentTypes {
		panic("too many component types registered in this scope")
	}

	reflectType := reflect.TypeFor[T]()
	if _, ok := scope.registeredTypes[reflectType]; ok {
		panic("component type already registered in this scope")
	}

	result := ComponentType[T]{
		id:      internal.TypeID(len(scope.registeredTypes)),
		storage: internal.NewComponentStorage[T](),
	}

	scope.registeredTypes[reflectType] = result
	scope.componentTypes[result.id] = result

	return result
}

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
