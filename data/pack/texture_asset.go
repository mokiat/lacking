package pack

import (
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/asset/mdl"
)

func BuildTwoDTextureAsset(image *mdl.Image) asset.Texture {
	return asset.Texture{
		Width:  uint32(image.Width()),
		Height: uint32(image.Height()),
		Flags:  asset.TextureFlag2D | asset.TextureFlagMipmapping,
		Format: asset.TexelFormatRGBA8,
		Layers: []asset.TextureLayer{
			{
				Data: image.DataRGBA8(),
			},
		},
	}
}
