package shading

type OperationDetails struct {
	ParamsOffset uint32
	ParamsCount  uint32
	Type         OperationType
}

type OperationType uint8

const (
	OperationDefineVec1 OperationType = 1 + iota
	OperationDefineVec2
	OperationDefineVec3
	OperationDefineVec4

	OperationTypeAssignVec1
	OperationTypeAssignVec2
	OperationTypeAssignVec3
	OperationTypeAssignVec4

	OperationTypeForwardOutputColor
	OperationTypeForwardAlphaDiscard
)

type OperationIndex uint32
