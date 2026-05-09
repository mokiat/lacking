package internal

type CommandHeader struct {
	CommandType CommandType
}

const (
	CommandTypeNone CommandType = iota
	CommandTypeEndOfSequence
	CommandTypeEditEntity
	CommandTypeDeleteEntity
	CommandTypeAddComponent
	CommandTypeRemoveComponent
)

type CommandType uint32

type EditEntityCommand struct {
	EntityID ID
}

type DeleteEntityCommand struct {
	EntityID ID
}

type AddComponentCommand struct {
	DataOffset uint32
	TypeID     TypeID
}

type RemoveComponentCommand struct {
	TypeID TypeID
}
