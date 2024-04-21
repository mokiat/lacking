package asset

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

// CullMode specifies the culling mode.
type CullMode uint8

const (
	// FaceOrientationCCW specifies that counter-clockwise primitives are
	// front-facing.
	FaceOrientationCCW FaceOrientation = iota

	// FaceOrientationCW specifies that clockwise primitives are front-facing.
	FaceOrientationCW
)

// FaceOrientation specifies the front face orientation.
type FaceOrientation uint8

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

// Comparison specifies the comparison function.
type Comparison uint8

const (
	// WrapModeClamp indicates that the texture coordinates should
	// be clamped to the range [0, 1].
	WrapModeClamp WrapMode = iota

	// WrapModeRepeat indicates that the texture coordinates should
	// be repeated.
	WrapModeRepeat

	// WrapModeMirroredRepeat indicates that the texture coordinates
	// should be repeated with mirroring.
	WrapModeMirroredRepeat
)

// WrapMode is an enumeration of the supported texture wrapping
// modes.
type WrapMode uint8

const (
	// FilterModeNearest indicates that the nearest texel should be
	// used for sampling.
	FilterModeNearest FilterMode = iota

	// FilterModeLinear indicates that the linear interpolation of
	// the nearest texels should be used for sampling.
	FilterModeLinear

	// FilterModeAnisotropic indicates that the anisotropic filtering
	// should be used for sampling.
	FilterModeAnisotropic
)

// FilterMode is an enumeration of the supported texture filtering
// modes.
type FilterMode uint8
