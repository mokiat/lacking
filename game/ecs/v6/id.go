package ecs

import "github.com/mokiat/lacking/game/ecs/v6/internal"

// NilID is the zero value of [ID] and represents an invalid entity handle.
var NilID = ID{}

// ID is a versioned handle to an entity. It remains valid until the
// entity is deleted. After deletion, the ID compares unequal to any
// new entity that reuses the same internal slot.
//
// ID is small and should be stored and passed by value.
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
