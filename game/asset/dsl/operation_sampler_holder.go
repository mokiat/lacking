package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// BindSampler sets the sampler of the target.
func BindSampler(name string, samplerProvider Provider[*mdl.Sampler]) Operation {
	type samplerHolder interface {
		SetSampler(name string, sampler *mdl.Sampler)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			sampler, err := samplerProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting sampler: %w", err)
			}

			container, ok := target.(samplerHolder)
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
