package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

// AddSkyLayer adds a sky layer to the target.
func AddSkyLayer(layerProvider Provider[mdl.SkyLayer]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			layer, err := layerProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting sky layer: %w", err)
			}

			sky, ok := target.(mdl.SkyLayerable)
			if !ok {
				return fmt.Errorf("target %T is not a sky layerable", target)
			}
			sky.AddSkyLayer(layer)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("add-sky-layer", layerProvider)
		},
	)
}

// SetSkyColor sets the color of the sky.
func SetSkyColor(colorProvider Provider[dprec.Vec4]) Operation {
	return SetProperty("skyColor", DPVec4ToSPVec4(colorProvider))
}

// SetSkyColorSampler sets the color sampler of the sky.
func SetSkyColorSampler(samplerProvider Provider[*mdl.Sampler]) Operation {
	return SetSampler("skyColorSampler", samplerProvider)
}
