package asset

const (
	TexelFormatR8 TexelFormat = iota
	TexelFormatR16
	TexelFormatR16F
	TexelFormatR32F
	TexelFormatRG8
	TexelFormatRG16
	TexelFormatRG16F
	TexelFormatRG32F
	TexelFormatRGB8
	TexelFormatRGB16
	TexelFormatRGB16F
	TexelFormatRGB32F
	TexelFormatRGBA8
	TexelFormatRGBA16
	TexelFormatRGBA16F
	TexelFormatRGBA32F
	TexelFormatDepth16F
	TexelFormatDepth32F
)

type TexelFormat uint8

const (
	TextureFlagNone       TextureFlag = 0
	TextureFlagMipmapping TextureFlag = 1 << iota
	TextureFlagLinearSpace
	TextureFlag2D
	TextureFlag2DArray
	TextureFlag3D
	TextureFlagCubeMap
)

type TextureFlag uint8

func (f TextureFlag) Has(flag TextureFlag) bool {
	return f&flag == flag
}

type Texture struct {
	Width  uint32
	Height uint32
	Format TexelFormat
	Flags  TextureFlag
	Layers []TextureLayer
}

type TextureLayer struct {
	Data []byte
}
