package internal

type CommandHeader struct {
	CommandType CommandType
}

const (
	CommandTypeNone CommandType = iota
	CommandTypeEditEntityBegin
	CommandTypeEditEntityEnd
	CommandTypeAddComponent
	CommandTypeRemoveComponent
)

type CommandType uint32

type EditEntityBeginCommand struct {
	EntityID ID
}

type AddComponentCommand struct {
	DataOffset uint32
	TypeID     TypeID
}

type RemoveComponentCommand struct {
	TypeID TypeID
}
