package pack

import (
	newasset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

func BuildTwoDTextureAsset(image *mdl.Image) newasset.Texture {
	return newasset.Texture{
		Width:  uint32(image.Width()),
		Height: uint32(image.Height()),
		Flags:  newasset.TextureFlag2D | newasset.TextureFlagMipmapping,
		Format: newasset.TexelFormatRGBA8,
		Layers: []newasset.TextureLayer{
			{
				Data: image.DataRGBA8(),
			},
		},
	}
}
