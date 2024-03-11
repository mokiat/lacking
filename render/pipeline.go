package render

// PipelineMarker marks a type as being a Pipeline.
type PipelineMarker interface {
	_isPipelineType()
}

// Pipeline is used to configure the GPU for drawing.
type Pipeline interface {
	PipelineMarker
	Resource
}

// PipelineInfo describes the information needed to create a new Pipeline.
type PipelineInfo struct {

	// Program specifies the shading to use.
	Program Program

	// VertexArray specifies the mesh data.
	VertexArray VertexArray

	// Topology specifies the primitive topology.
	Topology Topology

	// Culling specifies the culling mode.
	Culling CullMode

	// FrontFace specifies the front face orientation.
	FrontFace FaceOrientation

	// DepthTest specifies whether depth testing should be enabled.
	DepthTest bool

	// DepthWrite specifies whether depth writing should be enabled.
	DepthWrite bool

	// DepthComparison specifies the depth comparison function.
	DepthComparison Comparison

	// StencilTest specifies whether stencil testing should be enabled.
	StencilTest bool

	// StencilFront specifies the stencil operation state for front-facing
	// primitives.
	StencilFront StencilOperationState

	// StencilBack specifies the stencil operation state for back-facing
	// primitives.
	StencilBack StencilOperationState

	// ColorWrite specifies which color channels should be written to.
	ColorWrite [4]bool

	// BlendEnabled specifies whether blending should be enabled.
	BlendEnabled bool

	// BlendColor specifies the constant color that should be used for blending.
	BlendColor [4]float32

	// BlendSourceColorFactor specifies the source color factor for blending.
	BlendSourceColorFactor BlendFactor

	// BlendDestinationColorFactor specifies the destination color factor for
	// blending.
	BlendDestinationColorFactor BlendFactor

	// BlendSourceAlphaFactor specifies the source alpha factor for blending.
	BlendSourceAlphaFactor BlendFactor

	// BlendDestinationAlphaFactor specifies the destination alpha factor for
	// blending.
	BlendDestinationAlphaFactor BlendFactor

	// BlendOpColor specifies the color blend operation.
	BlendOpColor BlendOperation

	// BlendOpAlpha specifies the alpha blend operation.
	BlendOpAlpha BlendOperation
}

// Topology specifies the primitive topology.
type Topology uint8

const (
	// TopologyPoints specifies that the primitive topology is points.
	TopologyPoints Topology = iota

	// TopologyLineList specifies that the primitive topology is a line list.
	TopologyLineList

	// TopologyLineStrip specifies that the primitive topology is a line strip.
	TopologyLineStrip

	// TopologyTriangleList specifies that the primitive topology is a triangle
	// list.
	TopologyTriangleList

	// TopologyTriangleStrip specifies that the primitive topology is a triangle
	// strip.
	TopologyTriangleStrip

	// TopologyTriangleFan specifies that the primitive topology is a triangle
	// fan.
	//
	// TODO: This topology is not supported by WebGPU. Try to phase it out.
	TopologyTriangleFan
)

// CullMode specifies the culling mode.
type CullMode uint8

const (
	// CullModeNone specifies that no culling should be performed.
	CullModeNone CullMode = iota

	// CullModeFront specifies that front-facing primitives should be culled.
	CullModeFront

	// CullModeBack specifies that back-facing primitives should be culled.
	CullModeBack

	// CullModeFrontAndBack specifies that all oriented primitives should be
	// culled.
	CullModeFrontAndBack
)

// FaceOrientation specifies the front face orientation.
type FaceOrientation uint8

const (
	// FaceOrientationCCW specifies that counter-clockwise primitives are
	// front-facing.
	FaceOrientationCCW FaceOrientation = iota

	// FaceOrientationCW specifies that clockwise primitives are front-facing.
	FaceOrientationCW
)

// Comparison specifies the comparison function.
type Comparison uint8

const (
	// ComparisonNever specifies that the comparison should never pass.
	ComparisonNever Comparison = iota

	// ComparisonLess specifies that the comparison should pass if the source
	// value is less than the destination value.
	ComparisonLess

	// ComparisonEqual specifies that the comparison should pass if the source
	// value is equal to the destination value.
	ComparisonEqual

	// ComparisonLessOrEqual specifies that the comparison should pass if the
	// source value is less than or equal to the destination value.
	ComparisonLessOrEqual

	// ComparisonGreater specifies that the comparison should pass if the source
	// value is greater than the destination value.
	ComparisonGreater

	// ComparisonNotEqual specifies that the comparison should pass if the
	// source value is not equal to the destination value.
	ComparisonNotEqual

	// ComparisonGreaterOrEqual specifies that the comparison should pass if the
	// source value is greater than or equal to the destination value.
	ComparisonGreaterOrEqual

	// ComparisonAlways specifies that the comparison should always pass.
	ComparisonAlways
)

// StencilOperationState specifies the stencil operation state.
type StencilOperationState struct {

	// StencilFailOp specifies the operation to perform when the stencil test
	// fails.
	StencilFailOp StencilOperation

	// DepthFailOp specifies the operation to perform when the stencil test
	// passes, but the depth test fails.
	DepthFailOp StencilOperation

	// PassOp specifies the operation to perform when both the stencil test and
	// the depth test pass.
	PassOp StencilOperation

	// Comparison specifies the comparison function.
	Comparison Comparison

	// ComparisonMask specifies the comparison mask.
	ComparisonMask uint32

	// Reference specifies the reference value.
	Reference int32

	// WriteMask specifies the write mask.
	WriteMask uint32
}

// StencilOperation specifies the stencil operation.
type StencilOperation uint8

const (
	// StencilOperationKeep specifies that the current stencil value should be
	// kept.
	StencilOperationKeep StencilOperation = iota

	// StencilOperationZero specifies that the stencil value should be set to
	// zero.
	StencilOperationZero

	// StencilOperationReplace specifies that the stencil value should be set to
	// the reference value.
	StencilOperationReplace

	// StencilOperationIncrease specifies that the stencil value should be
	// incremented, clamping to the maximum value.
	StencilOperationIncrease

	// StencilOperationIncreaseWrap specifies that the stencil value should be
	// incremented, wrapping to zero if the maximum value is exceeded.
	StencilOperationIncreaseWrap

	// StencilOperationDecrease specifies that the stencil value should be
	// decremented, clamping to zero.
	StencilOperationDecrease

	// StencilOperationDecreaseWrap specifies that the stencil value should be
	// decremented, wrapping to the maximum value if zero is exceeded.
	StencilOperationDecreaseWrap

	// StencilOperationInvert specifies that the stencil value should be
	// bitwise inverted.
	StencilOperationInvert
)

var (
	// ColorMaskFalse specifies that no color channels should be written to.
	ColorMaskFalse = [4]bool{false, false, false, false}

	// ColorMaskTrue specifies that all color channels should be written to.
	ColorMaskTrue = [4]bool{true, true, true, true}
)

// BlendFactor specifies the blend factor.
type BlendFactor uint8

const (
	// BlendFactorZero specifies that the blend factor is zero.
	BlendFactorZero BlendFactor = iota

	// BlendFactorOne specifies that the blend factor is one.
	BlendFactorOne

	// BlendFactorSourceColor specifies that the blend factor is the source
	// color.
	BlendFactorSourceColor

	// BlendFactorOneMinusSourceColor specifies that the blend factor is one
	// minus the source color.
	BlendFactorOneMinusSourceColor

	// BlendFactorDestinationColor specifies that the blend factor is the
	// destination color.
	BlendFactorDestinationColor

	// BlendFactorOneMinusDestinationColor specifies that the blend factor is
	// one minus the destination color.
	BlendFactorOneMinusDestinationColor

	// BlendFactorSourceAlpha specifies that the blend factor is the source
	// alpha.
	BlendFactorSourceAlpha

	// BlendFactorOneMinusSourceAlpha specifies that the blend factor is one
	// minus the source alpha.
	BlendFactorOneMinusSourceAlpha

	// BlendFactorDestinationAlpha specifies that the blend factor is the
	// destination alpha.
	BlendFactorDestinationAlpha

	// BlendFactorOneMinusDestinationAlpha specifies that the blend factor is
	// one minus the destination alpha.
	BlendFactorOneMinusDestinationAlpha

	// BlendFactorConstantColor specifies that the blend factor is the constant
	// color.
	BlendFactorConstantColor

	// BlendFactorOneMinusConstantColor specifies that the blend factor is one
	// minus the constant color.
	BlendFactorOneMinusConstantColor

	// BlendFactorConstantAlpha specifies that the blend factor is the constant
	// alpha.
	BlendFactorConstantAlpha

	// BlendFactorOneMinusConstantAlpha specifies that the blend factor is one
	// minus the constant alpha.
	BlendFactorOneMinusConstantAlpha

	// BlendFactorSourceAlphaSaturate specifies that the blend factor is the
	// source alpha saturated.
	BlendFactorSourceAlphaSaturate
)

// BlendOperation specifies the blend operation.
type BlendOperation uint8

const (
	// BlendOperationAdd specifies that the blend operation is addition.
	BlendOperationAdd BlendOperation = iota

	// BlendOperationSubtract specifies that the blend operation is subtraction.
	BlendOperationSubtract

	// BlendOperationReverseSubtract specifies that the blend operation is
	// reverse subtraction.
	BlendOperationReverseSubtract

	// BlendOperationMin specifies that the blend operation is minimum.
	BlendOperationMin

	// BlendOperationMax specifies that the blend operation is maximum.
	BlendOperationMax
)
