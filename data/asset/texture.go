package asset

const (
	WrapModeDefault WrapMode = iota
	WrapModeRepeat
	WrapModeMirroredRepeat
	WrapModeClampToEdge
	WrapModeMirroredClampToEdge
)

type WrapMode uint8

const (
	FilterModeDefault FilterMode = iota
	FilterModeNearest
	FilterModeLinear
	FilterModeNearestMipmapNearest
	FilterModeNearestMipmapLinear
	FilterModeLinearMipmapNearest
	FilterModeLinearMipmapLinear
)

type FilterMode uint8

const (
	TexelFormatUnspecified TexelFormat = iota
	TexelFormatR8
	TexelFormatR16
	TexelFormatR32F
	TexelFormatRG8
	TexelFormatRG16
	TexelFormatRG32F
	TexelFormatRGB8
	TexelFormatRGB16
	TexelFormatRGB32F
	TexelFormatRGBA8
	TexelFormatRGBA16
	TexelFormatRGBA32F
	TexelFormatDepth32F
)

type TexelFormat uint8

const (
	TextureSideFront TextureSide = iota
	TextureSideBack
	TextureSideLeft
	TextureSideRight
	TextureSideTop
	TextureSideBottom
)

type TextureSide int

type TwoDTexture struct {
	Width     uint16
	Height    uint16
	WrapModeS WrapMode
	WrapModeT WrapMode
	MagFilter FilterMode
	MinFilter FilterMode
	Mipmaps   bool
	Format    TexelFormat
	Data      []byte
}

type CubeTextureSide struct {
	Data []byte
}

type CubeTexture struct {
	Dimension uint16
	WrapModeS WrapMode
	WrapModeT WrapMode
	WrapModeR WrapMode
	MagFilter FilterMode
	MinFilter FilterMode
	Mipmaps   bool
	Format    TexelFormat
	Sides     [6]CubeTextureSide
}
