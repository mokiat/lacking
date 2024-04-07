package render

// SamplerMarker marks a type as being a Sampler.
type SamplerMarker interface {
	_isSamplerType()
}

// Sampler represents a texture sampling configuration.
type Sampler interface {
	SamplerMarker
	Resource
}

// SamplerInfo represents the information needed to create a Sampler.
type SamplerInfo struct {

	// Wrapping specifies the texture wrapping mode.
	Wrapping WrapMode

	// Filtering specifies the texture filtering mode.
	Filtering FilterMode

	// Mipmapping specifies whether mipmapping should be enabled and whether
	// mipmaps should be generated.
	Mipmapping bool
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

// String returns a string representation of the WrapMode.
func (m WrapMode) String() string {
	switch m {
	case WrapModeClamp:
		return "CLAMP"
	case WrapModeRepeat:
		return "REPEAT"
	case WrapModeMirroredRepeat:
		return "MIRRORED_REPEAT"
	default:
		return "UNKNOWN"
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

// String returns a string representation of the FilterMode.
func (m FilterMode) String() string {
	switch m {
	case FilterModeNearest:
		return "NEAREST"
	case FilterModeLinear:
		return "LINEAR"
	case FilterModeAnisotropic:
		return "ANISOTROPIC"
	default:
		return "UNKNOWN"
	}
}
