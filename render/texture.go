package render

import (
	"fmt"

	"github.com/mokiat/gog/opt"
)

// TextureMarker marks a type as being a Texture.
type TextureMarker interface {
	_isTextureType()
}

// Texture is the interface that provides the API with the ability
// to store image data.
type Texture interface {
	TextureMarker

	// Release releases the resources associated with this Texture.
	Release()
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
type WrapMode int

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
		return fmt.Sprintf("UNKNOWN(%d)", m)
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
type FilterMode int

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
		return fmt.Sprintf("UNKNOWN(%d)", m)
	}
}

const (
	// DataFormatUnsupported indicates that the format is not supported.
	DataFormatUnsupported DataFormat = iota

	// DataFormatRGBA8 indicates that the format is RGBA8.
	DataFormatRGBA8

	// DataFormatRGBA16F indicates that the format is RGBA16F.
	DataFormatRGBA16F

	// DataFormatRGBA32F indicates that the format is RGBA32F.
	DataFormatRGBA32F
)

// DataFormat describes the format of the data that is stored in a
// Texture object.
type DataFormat int

// String returns a string representation of the DataFormat.
func (f DataFormat) String() string {
	switch f {
	case DataFormatUnsupported:
		return "UNSUPPORTED"
	case DataFormatRGBA8:
		return "RGBA8"
	case DataFormatRGBA16F:
		return "RGBA16F"
	case DataFormatRGBA32F:
		return "RGBA32F"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", f)
	}
}

// ColorTexture2DInfo represents the information needed to create a
// 2D color Texture.
type ColorTexture2DInfo struct {

	// Width specifies the width of the texture.
	Width int

	// Height specifies the height of the texture.
	Height int

	// Wrapping specifies the texture wrapping mode.
	Wrapping WrapMode

	// Filtering specifies the texture filtering mode.
	Filtering FilterMode

	// Mipmapping specifies whether mipmapping should be enabled and whether
	// mipmaps should be generated.
	Mipmapping bool

	// GammaCorrection specifies whether gamma correction should be performed
	// in order to convert the colors into linear space.
	GammaCorrection bool

	// Format specifies the format of the data.
	Format DataFormat

	// Data specifies the data that should be uploaded to the texture.
	Data []byte
}

// ColorTextureCubeInfo represents the information needed to create a
// cube color Texture.
type ColorTextureCubeInfo struct {

	// Dimension specifies the width, height and length of the texture.
	Dimension int

	// Filtering specifies the texture filtering mode.
	Filtering FilterMode

	// Mipmapping specifies whether mipmapping should be enabled and whether
	// mipmaps should be generated.
	Mipmapping bool

	// GammaCorrection specifies whether gamma correction should be performed
	// in order to convert the colors into linear space.
	GammaCorrection bool

	// Format specifies the format of the data.
	Format DataFormat

	// FrontSideData specifies the data that should be uploaded to the front
	// side of the texture.
	FrontSideData []byte

	// BackSideData specifies the data that should be uploaded to the back
	// side of the texture.
	BackSideData []byte

	// LeftSideData specifies the data that should be uploaded to the left
	// side of the texture.
	LeftSideData []byte

	// RightSideData specifies the data that should be uploaded to the right
	// side of the texture.
	RightSideData []byte

	// TopSideData specifies the data that should be uploaded to the top
	// side of the texture.
	TopSideData []byte

	// BottomSideData specifies the data that should be uploaded to the bottom
	// side of the texture.
	BottomSideData []byte
}

// DepthTexture2DInfo represents the information needed to create a
// 2D depth Texture.
type DepthTexture2DInfo struct {

	// Width specifies the width of the texture.
	Width int

	// Height specifies the height of the texture.
	Height int

	// ClippedValue specifies the value that should be used for depth clipping.
	ClippedValue opt.T[float32]

	// Comparable specifies whether the depth texture should be comparable.
	Comparable bool
}

// StencilTexture2DInfo represents the information needed to create a
// 2D stencil Texture.
type StencilTexture2DInfo struct {

	// Width specifies the width of the texture.
	Width int

	// Height specifies the height of the texture.
	Height int
}

// DepthStencilTexture2DInfo represents the information needed to create a
// 2D depth-stencil Texture.
type DepthStencilTexture2DInfo struct {

	// Width specifies the width of the texture.
	Width int

	// Height specifies the height of the texture.
	Height int

	// DepthClippedValue specifies the value that should be used for depth
	// clipping.
	DepthClippedValue opt.T[float32]
}
