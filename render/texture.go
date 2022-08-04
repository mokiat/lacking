package render

import "fmt"

type WrapMode int

const (
	WrapModeClamp WrapMode = iota
	WrapModeRepeat
	WrapModeMirroredRepeat
)

type FilterMode int

const (
	FilterModeNearest FilterMode = iota
	FilterModeLinear
	FilterModeAnisotropic
)

type DataFormat int

const (
	DataFormatUnsupported DataFormat = iota
	DataFormatRGBA8
	DataFormatRGBA16F
	DataFormatRGBA32F
)

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

type TextureObject interface {
	_isTextureObject() bool // ensures interface uniqueness
}

type Texture interface {
	TextureObject
	Release()
}

type ColorTexture2DInfo struct {
	Width           int
	Height          int
	Wrapping        WrapMode
	Filtering       FilterMode
	Mipmapping      bool
	GammaCorrection bool
	Format          DataFormat
	Data            []byte
}

type ColorTextureCubeInfo struct {
	Dimension       int
	Filtering       FilterMode
	Mipmapping      bool
	GammaCorrection bool
	Format          DataFormat
	FrontSideData   []byte
	BackSideData    []byte
	LeftSideData    []byte
	RightSideData   []byte
	TopSideData     []byte
	BottomSideData  []byte
}

type DepthTexture2DInfo struct {
	Width        int
	Height       int
	ClippedValue *float32
}

type StencilTexture2DInfo struct {
	Width  int
	Height int
}

type DepthStencilTexture2DInfo struct {
	Width             int
	Height            int
	DepthClippedValue *float32
}
