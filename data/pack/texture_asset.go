package pack

import (
	newasset "github.com/mokiat/lacking/game/newasset"
)

// TODO: Use Image model from mdl
func BuildTwoDTextureAsset(image *Image) newasset.Texture {
	return newasset.Texture{
		Width:  uint32(image.Width),
		Height: uint32(image.Height),
		Flags:  newasset.TextureFlag2D | newasset.TextureFlagMipmapping,
		Format: newasset.TexelFormatRGBA8,
		Layers: []newasset.TextureLayer{
			{
				Data: image.RGBA8Data(),
			},
		},
	}
}
