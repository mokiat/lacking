package ecs

import (
	"reflect"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/ecs/v5/internal"
)

// NewScope creates a new scope for component type registration.
func NewScope() *Scope {
	return &Scope{
		registeredTypes: ds.EmptySet[reflect.Type](),
		registry:        internal.NewRegistry(),
	}
}

// Scope represents a scope for component type registration.
type Scope struct {
	registeredTypes *ds.Set[reflect.Type]
	registry        *internal.Registry
}

// Type register the specified Go structure as a component type within
// the specified scope and returns its unique identifier. The identifier is used
// to refer to the component type in various API calls. It also acts as a
// storage for the component's data.
//
// This function should be called from a global variable initializer, and
// is not safe for concurrent use.
func Type[T any](scope *Scope) ComponentType[T] {
	initialCount := scope.registeredTypes.Size()

	if initialCount >= internal.MaxComponentTypes {
		panic("too many component types registered in this scope")
	}

	reflectType := reflect.TypeFor[T]()
	if scope.registeredTypes.Contains(reflectType) {
		panic("component type already registered in this scope")
	}
	scope.registeredTypes.Add(reflectType)

	id := internal.TypeID(initialCount)

	storage := internal.NewStorage[T]()
	scope.registry.SetStorage(id, storage)

	return ComponentType[T]{
		id:      id,
		storage: storage,
	}
}

// BaseComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type BaseComponentType interface {

	// BaseStorage returns the component storage associated with the component
	// type.
	BaseStorage() internal.BaseStorage
}

// ComponentType represents a component type in the ECS. It is used to
// identify component types and to manage component storage.
type ComponentType[T any] struct {
	id      internal.TypeID
	storage *internal.Storage[T]
}

var _ BaseComponentType = (*ComponentType[any])(nil)

// BaseStorage returns the component storage associated with the component
// type.
func (t ComponentType[T]) BaseStorage() internal.BaseStorage {
	return t.storage
}
