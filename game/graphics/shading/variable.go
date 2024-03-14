package shading

import "math"

type VariableDetails struct {
	Data   uint64
	Type   VariableType
	Source VariableSource
}

func (d VariableDetails) IsFloat32() bool {
	return d.Type == VariableTypeFloat32Value
}

func (d VariableDetails) AsFloat32() float32 {
	return math.Float32frombits(uint32(d.Data))
}

func (d VariableDetails) IsFloat64() bool {
	return d.Type == VariableTypeFloat64Value
}

func (d VariableDetails) AsFloat64() float64 {
	return math.Float64frombits(d.Data)
}

type VariableType uint8

const (
	VariableTypeFloat32Value VariableType = 1 + iota
	VariableTypeFloat64Value
	VariableTypeVec1
	VariableTypeVec2
	VariableTypeVec3
	VariableTypeVec4
)

type VariableSource uint8

const (
	VariableSourceCode VariableSource = 1 + iota
	VariableSourceUniform
	VariableSourceVarying
	VariableSourceSampler
)

type VariableIndex uint32

func (i VariableIndex) Index() VariableIndex {
	return i
}

type Vec1Variable struct {
	VariableIndex
}

type Vec2Variable struct {
	VariableIndex
}

type Vec3Variable struct {
	VariableIndex
}

type Vec4Variable struct {
	VariableIndex
}
