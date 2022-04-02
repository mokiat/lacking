package internal

import (
	"image"
	"image/draw"

	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui"
)

func NewImage(texture render.Texture, size ui.Size) *Image {
	return &Image{
		texture: texture,
		size:    size,
	}
}

type Image struct {
	texture render.Texture
	size    ui.Size
}

func (i *Image) Size() ui.Size {
	return i.size
}

func (i *Image) Destroy() {
	i.texture.Release()
}

func ImgToRGBA8(img image.Image) []byte {
	bounds := img.Bounds()
	var rgbaImg *image.NRGBA
	switch img := img.(type) {
	case *image.NRGBA:
		rgbaImg = img
	default:
		rgbaImg = image.NewNRGBA(bounds)
		draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)
	}
	return rgbaImg.Pix
}
