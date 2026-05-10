package ecs

import (
	"reflect"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/ecs/internal"
)

// NewScope creates a new component type registry.
//
// Once a scope is passed to [NewScene] it is locked; attempting to
// register additional component types afterwards will panic.
func NewScope() *Scope {
	return &Scope{
		registeredTypes: ds.EmptySet[reflect.Type](),
		registry:        internal.NewRegistry(),
	}
}

// Scope holds the component type registry shared by a set of scenes.
// Create one with [NewScope] and register component types with [Type].
type Scope struct {
	registeredTypes *ds.Set[reflect.Type]
	registry        *internal.Registry
	inUse           bool
}

func (s *Scope) markInUse() {
	s.inUse = true
}

// Type registers the Go type T as a component type within scope and
// returns a [ComponentType] descriptor. The descriptor is used in API
// calls such as [AddComponent], [RemoveComponent], and [GetComponent].
//
// Call this function once per type, typically from a package-level var
// initializer. It is not safe for concurrent use.
func Type[T any](scope *Scope) ComponentType[T] {
	if scope.inUse {
		panic("cannot register component type in a scope that is already in use")
	}

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

// ComponentType is the typed descriptor for a component of type T.
// Obtain one by calling [Type] and pass it to API functions such as
// [AddComponent], [GetComponent], and condition helpers like
// [HasComponent].
type ComponentType[T any] struct {
	id      internal.TypeID
	storage *internal.Storage[T]
}
