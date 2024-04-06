package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func SetCubeSide(side mdl.CubeSide, imageProvider Provider[*mdl.Image]) Operation {
	apply := func(target any) error {
		image, err := imageProvider.Get()
		if err != nil {
			return fmt.Errorf("error getting image: %w", err)
		}

		texture, ok := target.(*mdl.Texture)
		if !ok {
			return fmt.Errorf("target %T is not a texture", target)
		}
		texture.SetLayerImage(int(side), image)

		return nil
	}

	digest := func() ([]byte, error) {
		return digestItems("set-cube-side", uint8(side), imageProvider)
	}

	return FuncOperation(apply, digest)
}
