package shading

type OperationDetails struct {
	InputParams  ParamRange
	OutputParams ParamRange
	Type         OperationType
}

type ParamRange struct {
	Offset uint32
	Count  uint32
}

type OperationType uint16

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
