package ecs

import (
	"reflect"
)

// MaxComponentTypes returns the maximum number of component types that can be
// registered within a scope.
const MaxComponentTypes = 512

// NewScope creates a new scope for component type registration.
func NewScope() *Scope {
	return &Scope{
		registeredTypes: make(map[reflect.Type]BaseComponentType),
	}
}

// Scope represents a scope for component type registration.
type Scope struct {
	registeredTypes map[reflect.Type]BaseComponentType
	componentTypes  [MaxComponentTypes]BaseComponentType
}

func (s *Scope) getComponentTypeByID(id typeID) BaseComponentType {
	return s.componentTypes[id]
}

// RegisterType register the specified Go structure as a component type within
// the specified scope and returns its unique identifier. The identifier is used
// to refer to the component type in various API calls. It also acts as a
// storage for the component's data.
//
// This function should be called from a global variable initializer, and
// is not safe for concurrent use.
func RegisterType[T any](scope *Scope) ComponentType[T] {
	if len(scope.registeredTypes) >= MaxComponentTypes {
		panic("too many component types registered in this scope")
	}

	reflectType := reflect.TypeFor[T]()
	if _, ok := scope.registeredTypes[reflectType]; ok {
		panic("component type already registered in this scope")
	}

	tIndex := typeID(len(scope.registeredTypes))
	result := newComponentType[T](tIndex)
	scope.registeredTypes[reflectType] = result
	scope.componentTypes[tIndex] = result

	return result
}
