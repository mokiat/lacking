package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

// CreateSampler creates a new sampler with the provided texture and operations.
func CreateSampler(textureProvider Provider[*mdl.Texture], operations ...Operation) Provider[*mdl.Sampler] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Sampler, error) {
			texture, err := textureProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting texture: %w", err)
			}

			var sampler mdl.Sampler
			sampler.SetWrapMode(mdl.WrapModeClamp)
			sampler.SetFilterMode(mdl.FilterModeNearest)
			sampler.SetMipmapping(false)
			sampler.SetTexture(texture)
			for _, op := range operations {
				if err := op.Apply(&sampler); err != nil {
					return nil, err
				}
			}
			return &sampler, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-sampler", textureProvider, operations)
		},
	))
}
