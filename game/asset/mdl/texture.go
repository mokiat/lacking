package mdl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
)

const (
	TextureFormatR8       TextureFormat = dto.TexelFormatR8
	TextureFormatR16      TextureFormat = dto.TexelFormatR16
	TextureFormatR16F     TextureFormat = dto.TexelFormatR16F
	TextureFormatR32F     TextureFormat = dto.TexelFormatR32F
	TextureFormatRG8      TextureFormat = dto.TexelFormatRG8
	TextureFormatRG16     TextureFormat = dto.TexelFormatRG16
	TextureFormatRG16F    TextureFormat = dto.TexelFormatRG16F
	TextureFormatRG32F    TextureFormat = dto.TexelFormatRG32F
	TextureFormatRGB8     TextureFormat = dto.TexelFormatRGB8
	TextureFormatRGB16    TextureFormat = dto.TexelFormatRGB16
	TextureFormatRGB16F   TextureFormat = dto.TexelFormatRGB16F
	TextureFormatRGB32F   TextureFormat = dto.TexelFormatRGB32F
	TextureFormatRGBA8    TextureFormat = dto.TexelFormatRGBA8
	TextureFormatRGBA16   TextureFormat = dto.TexelFormatRGBA16
	TextureFormatRGBA16F  TextureFormat = dto.TexelFormatRGBA16F
	TextureFormatRGBA32F  TextureFormat = dto.TexelFormatRGBA32F
	TextureFormatDepth16F TextureFormat = dto.TexelFormatDepth16F
	TextureFormatDepth32F TextureFormat = dto.TexelFormatDepth32F
)

type TextureFormat = dto.TexelFormat

const (
	TextureKind2D TextureKind = iota
	TextureKind2DArray
	TextureKind3D
	TextureKindCube
)

type TextureKind uint8

func Create2DTexture(width, height, mipmaps int, format TextureFormat) *Texture {
	mipmapLayers := make([]MipmapLayer, mipmaps)
	for i := range mipmapLayers {
		mipWidth := max(1, width>>i)
		mipHeight := max(1, height>>i)
		mipTexelSize := textureFormatSize(format)
		mipmapLayers[i] = MipmapLayer{
			width:  mipWidth,
			height: mipHeight,
			depth:  1,
			layers: []TextureLayer{
				{
					data: make([]byte, mipWidth*mipHeight*mipTexelSize),
				},
			},
		}
	}
	return &Texture{
		Object:       NewObject(),
		kind:         TextureKind2D,
		format:       format,
		mipmapLayers: mipmapLayers,
	}
}

func CreateCubeTexture(dimension, mipmaps int, format TextureFormat) *Texture {
	mipmapLayers := make([]MipmapLayer, mipmaps)
	for i := range mipmapLayers {
		mipDimension := max(1, dimension>>i)
		mipTexelSize := textureFormatSize(format)
		mipmapLayers[i] = MipmapLayer{
			width:  mipDimension,
			height: mipDimension,
			depth:  1,
			layers: []TextureLayer{
				{
					data: make([]byte, mipDimension*mipDimension*mipTexelSize),
				},
				{
					data: make([]byte, mipDimension*mipDimension*mipTexelSize),
				},
				{
					data: make([]byte, mipDimension*mipDimension*mipTexelSize),
				},
				{
					data: make([]byte, mipDimension*mipDimension*mipTexelSize),
				},
				{
					data: make([]byte, mipDimension*mipDimension*mipTexelSize),
				},
				{
					data: make([]byte, mipDimension*mipDimension*mipTexelSize),
				},
			},
		}
	}
	return &Texture{
		Object:       NewObject(),
		kind:         TextureKindCube,
		format:       format,
		mipmapLayers: mipmapLayers,
	}
}

type Texture struct {
	*Object
	name            string
	kind            TextureKind
	format          TextureFormat
	generateMipmaps bool
	isLinear        bool
	mipmapLayers    []MipmapLayer
}

func (t *Texture) Name() string {
	return t.name
}

func (t *Texture) SetName(name string) {
	t.name = name
}

func (t *Texture) Kind() TextureKind {
	return t.kind
}

func (t *Texture) Format() TextureFormat {
	return t.format
}

func (t *Texture) Linear() bool {
	return t.isLinear
}

func (t *Texture) SetLinear(isLinear bool) {
	t.isLinear = isLinear
}

func (t *Texture) GenerateMipmaps() bool {
	return t.generateMipmaps
}

func (t *Texture) SetGenerateMipmaps(generateMipmaps bool) {
	t.generateMipmaps = generateMipmaps
}

func (t *Texture) SetLayerImage(mipmap, index int, image *Image) {
	mipmapLayer := t.mipmapLayers[mipmap]
	switch t.format {
	case TextureFormatRGBA8:
		copy(mipmapLayer.layers[index].data, image.DataRGBA8())
	case TextureFormatRGBA16F:
		copy(mipmapLayer.layers[index].data, image.DataRGBA16F())
	case TextureFormatRGBA32F:
		copy(mipmapLayer.layers[index].data, image.DataRGBA32F())
	default:
		panic(fmt.Errorf("unsupported texture format: %v", t.format))
	}
}

func (t *Texture) MipmapLayers() []MipmapLayer {
	return t.mipmapLayers
}

type MipmapLayer struct {
	width  int
	height int
	depth  int
	layers []TextureLayer
}

func (l *MipmapLayer) Width() int {
	return l.width
}

func (l *MipmapLayer) Height() int {
	return l.height
}

func (l *MipmapLayer) Depth() int {
	return l.depth
}

func (l *MipmapLayer) Layers() []TextureLayer {
	return l.layers
}

type TextureLayer struct {
	data []byte
}

func (l *TextureLayer) Data() []byte {
	return l.data
}

func textureFormatSize(format TextureFormat) int {
	switch format {
	case TextureFormatRGBA8:
		return 4
	case TextureFormatRGBA16F:
		return 8
	case TextureFormatRGBA32F:
		return 16
	default:
		panic(fmt.Errorf("unsupported texture format: %v", format))
	}
}
