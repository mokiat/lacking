package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// SetFormat configures the format of the target.
func SetFormat(formatProvider Provider[mdl.TextureFormat]) Operation {
	type formatConfigurable interface {
		SetFormat(mdl.TextureFormat)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			format, err := formatProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting format: %w", err)
			}

			configurable, ok := target.(formatConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not configurable with format", target)
			}
			configurable.SetFormat(format)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-format", formatProvider)
		},
	)
}
