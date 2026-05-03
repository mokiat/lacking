package ecs

// NilEntityID represents an invalid entity handle.
var NilEntityID = EntityID{}

// EntityID represents a handle to an ECS entity. The handle may be invalid
// if the entity has since been deleted.
//
// Store this type by value, as it is small and copyable.
type EntityID struct {
	index    int32
	revision int32
}
