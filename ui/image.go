package ui

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/mokiat/lacking/render"
)

func newImage(texture render.Texture, size Size) *Image {
	return &Image{
		texture: texture,
		size:    size,
	}
}

// Image represents a 2D image.
type Image struct {
	texture render.Texture
	size    Size
}

// Size returns the dimensions of this Image.
func (i *Image) Size() Size {
	return i.size
}

// Destroy releases all resources allocated for this
// image.
func (i *Image) Destroy() {
	fmt.Println("DESTROYING IMAGE")
	i.texture.Release()
}

func imgToRGBA8(img image.Image) []byte {
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
