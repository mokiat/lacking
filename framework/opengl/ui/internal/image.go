package internal

import (
	"image"
	"image/draw"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

func NewImage() *Image {
	return &Image{
		texture: opengl.NewTwoDTexture(),
		size:    ui.NewSize(0, 0),
	}
}

type Image struct {
	texture *opengl.TwoDTexture
	size    ui.Size
}

func (i *Image) Allocate(img image.Image) {
	bounds := img.Bounds()
	var rgbaImg *image.NRGBA
	switch img := img.(type) {
	case *image.NRGBA:
		rgbaImg = img
	default:
		rgbaImg = image.NewNRGBA(bounds)
		draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)
	}
	i.size = ui.NewSize(bounds.Dx(), bounds.Dy())
	i.texture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             int32(bounds.Dx()),
		Height:            int32(bounds.Dy()),
		WrapS:             gl.CLAMP_TO_EDGE,
		WrapT:             gl.CLAMP_TO_EDGE,
		MinFilter:         gl.LINEAR,
		MagFilter:         gl.LINEAR,
		UseAnisotropy:     false,
		GenerateMipmaps:   false,
		InternalFormat:    gl.SRGB8_ALPHA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
		Data:              rgbaImg.Pix,
	})
}

func (i *Image) Release() {
	i.size = ui.NewSize(0, 0)
	i.texture.Release()
}

func (i *Image) Size() ui.Size {
	return i.size
}
