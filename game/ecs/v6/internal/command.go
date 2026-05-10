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
	CommandTypeAddComponent
	CommandTypeRemoveComponent
	CommandTypeReplaceComponent
)

type CommandType uint32

type CreateEntityCommand struct {
	EntityID ID
}

type EditEntityCommand struct {
	EntityID ID
	StageRow Row
}

type DeleteEntityCommand struct {
	EntityID ID
}

type AddComponentCommand struct {
	TypeID TypeID
}

type RemoveComponentCommand struct {
	TypeID TypeID
}

type ReplaceComponentCommand struct {
	TypeID TypeID
}
