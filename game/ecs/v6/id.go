package ecs

import "github.com/mokiat/lacking/game/ecs/v6/internal"

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

func fromInternalID(internalID internal.ID) ID {
	return ID{
		index:    internalID.Index,
		revision: internalID.Revision,
	}
}
