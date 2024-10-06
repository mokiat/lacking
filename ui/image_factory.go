package ui

import (
	"image"

	"github.com/mokiat/lacking/render"
)

func newImageFactory(api render.API) *imageFactory {
	return &imageFactory{
		api: api,
	}
}

type imageFactory struct {
	api render.API
}

func (f *imageFactory) CreateImage(img image.Image) *Image {
	bounds := img.Bounds()
	size := NewSize(bounds.Dx(), bounds.Dy())
	texture := f.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		GenerateMipmaps: true,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
		MipmapLayers: []render.Mipmap2DLayer{
			{
				Width:  uint32(size.Width),
				Height: uint32(size.Height),
				Data:   imgToRGBA8(img),
			},
		},
	})
	return newImage(texture, size)
}
