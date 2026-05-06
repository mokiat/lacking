package ecs

// NilID represents an invalid entity handle.
var NilID = ID{}

// ID represents a handle to an ECS entity. The handle may be invalid
// if the entity has since been deleted.
//
// Store this type by value, as it is small and copyable.
type ID struct {
	index    uint32
	revision uint32
}
