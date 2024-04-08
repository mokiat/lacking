package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func CreateCubeTexture(format mdl.TextureFormat, cubeImageProvider Provider[*mdl.CubeImage], operations ...Operation) Provider[*mdl.Texture] {
	get := func() (*mdl.Texture, error) {
		cubeImage, err := cubeImageProvider.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get cube image: %w", err)
		}

		frontImage := cubeImage.SideToImage(mdl.CubeSideFront)
		rearImage := cubeImage.SideToImage(mdl.CubeSideRear)
		leftImage := cubeImage.SideToImage(mdl.CubeSideLeft)
		rightImage := cubeImage.SideToImage(mdl.CubeSideRight)
		topImage := cubeImage.SideToImage(mdl.CubeSideTop)
		bottomImage := cubeImage.SideToImage(mdl.CubeSideBottom)

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
		for _, op := range operations {
			if err := op.Apply(&texture); err != nil {
				return nil, err
			}
		}
		return &texture, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("cube-texture", operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
