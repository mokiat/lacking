package internal

var NilID = ID{}

func NewID(index uint32, revision uint32) ID {
	return ID{
		Index:    index,
		Revision: revision,
	}
}

// ID represents a handle to an ECS entity. The handle may be invalid
// if the entity has since been deleted.
//
// Store this type by value, as it is small and copyable.
type ID struct {
	Index    uint32
	Revision uint32
}
