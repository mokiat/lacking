package internal

type CommandHeader struct {
	CommandType CommandType
}

const (
	CommandTypeNone CommandType = iota
	CommandTypeEndOfSequence
	CommandTypeCreateEntity
	CommandTypeEditEntity
	CommandTypeDeleteEntity
	CommandTypeSetComponent
	CommandTypeUnsetComponent
)

type CommandType uint32

type CreateEntityCommand struct {
	EntityID ID
	StageRow Row
}

type EditEntityCommand struct {
	EntityID ID
	StageRow Row
}

type DeleteEntityCommand struct {
	EntityID ID
}

type SetComponentCommand struct {
	TypeID TypeID
}

type UnsetComponentCommand struct {
	TypeID TypeID
}
