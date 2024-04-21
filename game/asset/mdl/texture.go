package mdl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
)

const (
	TextureFormatR8       TextureFormat = asset.TexelFormatR8
	TextureFormatR16      TextureFormat = asset.TexelFormatR16
	TextureFormatR16F     TextureFormat = asset.TexelFormatR16F
	TextureFormatR32F     TextureFormat = asset.TexelFormatR32F
	TextureFormatRG8      TextureFormat = asset.TexelFormatRG8
	TextureFormatRG16     TextureFormat = asset.TexelFormatRG16
	TextureFormatRG16F    TextureFormat = asset.TexelFormatRG16F
	TextureFormatRG32F    TextureFormat = asset.TexelFormatRG32F
	TextureFormatRGB8     TextureFormat = asset.TexelFormatRGB8
	TextureFormatRGB16    TextureFormat = asset.TexelFormatRGB16
	TextureFormatRGB16F   TextureFormat = asset.TexelFormatRGB16F
	TextureFormatRGB32F   TextureFormat = asset.TexelFormatRGB32F
	TextureFormatRGBA8    TextureFormat = asset.TexelFormatRGBA8
	TextureFormatRGBA16   TextureFormat = asset.TexelFormatRGBA16
	TextureFormatRGBA16F  TextureFormat = asset.TexelFormatRGBA16F
	TextureFormatRGBA32F  TextureFormat = asset.TexelFormatRGBA32F
	TextureFormatDepth16F TextureFormat = asset.TexelFormatDepth16F
	TextureFormatDepth32F TextureFormat = asset.TexelFormatDepth32F
)

type TextureFormat = asset.TexelFormat

const (
	TextureKind2D TextureKind = iota
	TextureKind2DArray
	TextureKind3D
	TextureKindCube
)

type TextureKind uint8

type Texture struct {
	kind   TextureKind
	width  int
	height int
	format TextureFormat
	layers []TextureLayer
}

func (t *Texture) Kind() TextureKind {
	return t.kind
}

func (t *Texture) SetKind(kind TextureKind) {
	t.kind = kind
}

func (t *Texture) Width() int {
	return t.width
}

func (t *Texture) Height() int {
	return t.height
}

func (t *Texture) Format() TextureFormat {
	return t.format
}

func (t *Texture) SetFormat(format TextureFormat) {
	t.format = format
	if len(t.layers) > 0 {
		panic("setting texture format with layers is not supported yet")
	}
}

func (t *Texture) Resize(width, height int) {
	t.width = width
	t.height = height
	if len(t.layers) > 0 {
		panic("resizing texture with layers is not supported yet")
	}
}

func (t *Texture) EnsureLayer(index int) {
	for index >= len(t.layers) {
		t.layers = append(t.layers, createTextureLayer(t.width, t.height, t.format))
	}
}

func (t *Texture) SetLayerImage(index int, image *Image) {
	t.EnsureLayer(index)

	if image.width != t.width || image.height != t.height {
		image = image.Scale(t.width, t.height)
	}
	switch t.format {
	case TextureFormatRGBA8:
		copy(t.layers[index].data, image.DataRGBA8())
	case TextureFormatRGBA16F:
		copy(t.layers[index].data, image.DataRGBA16F())
	case TextureFormatRGBA32F:
		copy(t.layers[index].data, image.DataRGBA32F())
	default:
		panic(fmt.Errorf("unsupported texture format: %v", t.format))
	}
}

func createTextureLayer(width, height int, format TextureFormat) TextureLayer {
	var texelSize int
	switch format {
	case TextureFormatRGBA8:
		texelSize = 4
	case TextureFormatRGBA16F:
		texelSize = 8
	case TextureFormatRGBA32F:
		texelSize = 16
	default:
		panic(fmt.Errorf("unsupported texture format: %v", format))
	}
	return TextureLayer{
		data: make([]byte, width*height*texelSize),
	}
}

type TextureLayer struct {
	data []byte
}

func (l *TextureLayer) Data() []byte {
	return l.data
}
