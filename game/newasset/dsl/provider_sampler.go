package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func CreateSampler(textureProvider Provider[*mdl.Texture], operations ...Operation) Provider[*mdl.Sampler] {
	get := func() (*mdl.Sampler, error) {
		texture, err := textureProvider.Get()
		if err != nil {
			return nil, fmt.Errorf("error getting texture: %w", err)
		}

		var sampler mdl.Sampler
		sampler.SetTexture(texture)
		for _, op := range operations {
			if err := op.Apply(&sampler); err != nil {
				return nil, err
			}
		}
		return &sampler, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("sampler", textureProvider, operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
