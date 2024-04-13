package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func SetSampler(name string, samplerProvider Provider[*mdl.Sampler]) Operation {
	apply := func(target any) error {
		sampler, err := samplerProvider.Get()
		if err != nil {
			return fmt.Errorf("error getting sampler: %w", err)
		}
		container, ok := target.(mdl.TextureHolder)
		if !ok {
			return fmt.Errorf("target %T is not a sampler holder", target)
		}
		container.SetSampler(name, sampler)
		return nil
	}

	digest := func() ([]byte, error) {
		return CreateDigest("set-sampler", name, samplerProvider)
	}

	return FuncOperation(apply, digest)
}
