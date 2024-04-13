package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

// SetSampler sets the sampler of the target.
func SetSampler(name string, samplerProvider Provider[*mdl.Sampler]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			sampler, err := samplerProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting sampler: %w", err)
			}

			container, ok := target.(mdl.SamplerHolder)
			if !ok {
				return fmt.Errorf("target %T is not a sampler holder", target)
			}
			container.SetSampler(name, sampler)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-sampler", name, samplerProvider)
		},
	)
}
