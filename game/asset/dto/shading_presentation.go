package dto

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

// String returns the string representation of the CullMode.
func (m CullMode) String() string {
	switch m {
	case CullModeNone:
		return "none"
	case CullModeFront:
		return "front"
	case CullModeBack:
		return "back"
	case CullModeFrontAndBack:
		return "front-and-back"
	default:
		return "unknown"
	}
}

const (
	// FaceOrientationCCW specifies that counter-clockwise primitives are
	// front-facing.
	FaceOrientationCCW FaceOrientation = iota

	// FaceOrientationCW specifies that clockwise primitives are front-facing.
	FaceOrientationCW
)

// FaceOrientation specifies the front face orientation.
type FaceOrientation uint8

// String returns the string representation of the FaceOrientation.
func (m FaceOrientation) String() string {
	switch m {
	case FaceOrientationCCW:
		return "ccw"
	case FaceOrientationCW:
		return "cw"
	default:
		return "unknown"
	}
}

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

// String returns the string representation of the Comparison.
func (m Comparison) String() string {
	switch m {
	case ComparisonNever:
		return "never"
	case ComparisonLess:
		return "less"
	case ComparisonEqual:
		return "equal"
	case ComparisonLessOrEqual:
		return "less-or-equal"
	case ComparisonGreater:
		return "greater"
	case ComparisonNotEqual:
		return "not-equal"
	case ComparisonGreaterOrEqual:
		return "greater-or-equal"
	case ComparisonAlways:
		return "always"
	default:
		return "unknown"
	}
}

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

// String returns the string representation of the WrapMode.
func (m WrapMode) String() string {
	switch m {
	case WrapModeClamp:
		return "clamp"
	case WrapModeRepeat:
		return "repeat"
	case WrapModeMirroredRepeat:
		return "mirrored-repeat"
	default:
		return "unknown"
	}
}

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

// String returns the string representation of the FilterMode.
func (m FilterMode) String() string {
	switch m {
	case FilterModeNearest:
		return "nearest"
	case FilterModeLinear:
		return "linear"
	case FilterModeAnisotropic:
		return "anisotropic"
	default:
		return "unknown"
	}
}
