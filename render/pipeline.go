package render

import "github.com/mokiat/gomath/sprec"

type PipelineInfo struct {
	Program                     Program
	VertexArray                 VertexArray
	Topology                    Topology
	Culling                     CullMode
	FrontFace                   FaceOrientation
	LineWidth                   float32
	DepthTest                   bool
	DepthWrite                  bool
	DepthComparison             Comparison
	StencilTest                 bool
	StencilFront                StencilOperationState
	StencilBack                 StencilOperationState
	ColorWrite                  [4]bool
	BlendEnabled                bool
	BlendColor                  sprec.Vec4
	BlendSourceColorFactor      BlendFactor
	BlendDestinationColorFactor BlendFactor
	BlendSourceAlphaFactor      BlendFactor
	BlendDestinationAlphaFactor BlendFactor
	BlendOpColor                BlendOperation
	BlendOpAlpha                BlendOperation
}

type Topology uint8

const (
	TopologyPoints Topology = iota
	TopologyLineStrip
	TopologyLineLoop
	TopologyLines
	TopologyTriangleStrip
	TopologyTriangleFan
	TopologyTriangles
)

type CullMode uint8

const (
	CullModeNone CullMode = iota
	CullModeFront
	CullModeBack
	CullModeFrontAndBack
)

type FaceOrientation uint8

const (
	FaceOrientationCCW FaceOrientation = iota
	FaceOrientationCW
)

type Comparison uint8

const (
	ComparisonNever Comparison = iota
	ComparisonLess
	ComparisonEqual
	ComparisonLessOrEqual
	ComparisonGreater
	ComparisonNotEqual
	ComparisonGreaterOrEqual
	ComparisonAlways
)

type StencilOperationState struct {
	StencilFailOp  StencilOperation
	DepthFailOp    StencilOperation
	PassOp         StencilOperation
	Comparison     Comparison
	ComparisonMask uint32
	Reference      uint32
	WriteMask      uint32
}

type StencilOperation uint8

const (
	StencilOperationKeep StencilOperation = iota
	StencilOperationZero
	StencilOperationReplace
	StencilOperationIncrease
	StencilOperationIncreaseWrap
	StencilOperationDecrease
	StencilOperationDecreaseWrap
	StencilOperationInvert
)

type BlendFactor uint8

const (
	BlendFactorZero BlendFactor = iota
	BlendFactorOne
	BlendFactorSourceColor
	BlendFactorOneMinusSourceColor
	BlendFactorDestinationColor
	BlendFactorOneMinusDestinationColor
	BlendFactorSourceAlpha
	BlendFactorOneMinusSourceAlpha
	BlendFactorDestinationAlpha
	BlendFactorOneMinusDestinationAlpha
	BlendFactorConstantColor
	BlendFactorOneMinusConstantColor
	BlendFactorConstantAlpha
	BlendFactorOneMinusConstantAlpha
	BlendFactorSourceAlphaSaturate
)

type BlendOperation uint8

const (
	BlendOperationAdd BlendOperation = iota
	BlendOperationSubtract
	BlendOperationReverseSubtract
	BlendOperationMin
	BlendOperationMax
)

type Pipeline interface {
	Release()
}