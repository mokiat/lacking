package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

// CreateCubeTexture creates a new cube texture with the specified format and
// source image.
func CreateCubeTexture(format mdl.TextureFormat, cubeImageProvider Provider[*mdl.CubeImage]) Provider[*mdl.Texture] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Texture, error) {
			cubeImage, err := cubeImageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get cube image: %w", err)
			}

			frontImage := cubeImage.Side(mdl.CubeSideFront)
			rearImage := cubeImage.Side(mdl.CubeSideRear)
			leftImage := cubeImage.Side(mdl.CubeSideLeft)
			rightImage := cubeImage.Side(mdl.CubeSideRight)
			topImage := cubeImage.Side(mdl.CubeSideTop)
			bottomImage := cubeImage.Side(mdl.CubeSideBottom)

			var texture mdl.Texture
			texture.SetKind(mdl.TextureKindCube)
			texture.SetFormat(format)
			texture.Resize(frontImage.Width(), frontImage.Height())
			texture.SetLayerImage(0, frontImage)
			texture.SetLayerImage(1, rearImage)
			texture.SetLayerImage(2, leftImage)
			texture.SetLayerImage(3, rightImage)
			texture.SetLayerImage(4, topImage)
			texture.SetLayerImage(5, bottomImage)
			return &texture, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-cube-texture", uint8(format), cubeImageProvider)
		},
	))
}

var defaultCubeTextureProvider = CreateCubeTexture(mdl.TextureFormatRGBA32F, defaultCubeImageProvider)
