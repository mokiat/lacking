package dsl

import (
	"fmt"
)

// WithSampleCount configures the sample count of the target.
func WithSampleCount(countProvider Provider[int]) Operation {
	type sampleCountConfigurable interface {
		SetSampleCount(int)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			count, err := countProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting sample count: %w", err)
			}

			configurable, ok := target.(sampleCountConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not configurable with sample count", target)
			}
			configurable.SetSampleCount(count)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("with-sample-count", countProvider)
		},
	)
}
