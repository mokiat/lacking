package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func AddSkyLayer(layerProvider Provider[mdl.SkyLayer]) Operation {
	apply := func(target any) error {
		sky, ok := target.(mdl.SkyLayerable)
		if !ok {
			return fmt.Errorf("target %T is not a sky layerable", target)
		}
		layer, err := layerProvider.Get()
		if err != nil {
			return fmt.Errorf("error getting sky layer: %w", err)
		}
		sky.AddSkyLayer(layer)
		return nil
	}

	digest := func() ([]byte, error) {
		return CreateDigest("add-sky-layer", layerProvider)
	}

	return FuncOperation(apply, digest)
}
