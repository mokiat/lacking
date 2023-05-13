package ecs

import "errors"

var freeComponentTypeID ComponentTypeID

// NewComponentTypeID creates a new unique ComponentTypeID.
func NewComponentTypeID() ComponentTypeID {
	if freeComponentTypeID >= 64 {
		panic(errors.New("max component type count reached"))
	}
	result := freeComponentTypeID
	freeComponentTypeID++
	return result
}

// ComponentTypeID is an identifier for a component type.
// Numbers should be in the range [0..63]. That is, the
// ECS framework supports at most 64 component types at
// the moment.
type ComponentTypeID uint8

func (i ComponentTypeID) mask() uint64 {
	return 1 << i
}

// Component represents an ECS component that can be attached to an Entity.
type Component interface {
	TypeID() ComponentTypeID
}

// AttachComponent is a helper generic function that allows one to attach
// a component to an Entity without dealing with type identifiers.
func AttachComponent[T Component](entity *Entity, component T) {
	typeID := component.TypeID()
	entity.SetComponent(typeID, component)
}

// FetchComponent is a helper generic function that allows one to check whether
// a given Entity has a particular component and if it does, the component will
// be injected into the specified pointer target.
func FetchComponent[T Component](entity *Entity, target *T) bool {
	if target == nil {
		return false
	}
	typeID := (*target).TypeID()
	comp := entity.Component(typeID)
	if comp == nil {
		return false
	}
	*target = comp.(T)
	return true
}
