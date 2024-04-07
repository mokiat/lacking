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
	Resource
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
type DataFormat uint8

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
	Width int // TODO: uint32

	// Height specifies the height of the texture.
	Height int // TODO: uint32

	// GenerateMipmaps specifies whether mipmaps should be generated.
	GenerateMipmaps bool

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
	Dimension int // TODO: uint32

	// GenerateMipmaps specifies whether mipmaps should be generated.
	GenerateMipmaps bool

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
	Width int // TODO: uint32

	// Height specifies the height of the texture.
	Height int // TODO: uint32

	// ClippedValue specifies the value that should be used for depth clipping.
	ClippedValue opt.T[float32]

	// Comparable specifies whether the depth texture should be comparable.
	Comparable bool
}

// StencilTexture2DInfo represents the information needed to create a
// 2D stencil Texture.
type StencilTexture2DInfo struct {

	// Width specifies the width of the texture.
	Width int // TODO: uint32

	// Height specifies the height of the texture.
	Height int // TODO: uint32
}

// DepthStencilTexture2DInfo represents the information needed to create a
// 2D depth-stencil Texture.
type DepthStencilTexture2DInfo struct {

	// Width specifies the width of the texture.
	Width int // TODO: uint32

	// Height specifies the height of the texture.
	Height int // TODO: uint32

	// DepthClippedValue specifies the value that should be used for depth
	// clipping.
	DepthClippedValue opt.T[float32]
}
