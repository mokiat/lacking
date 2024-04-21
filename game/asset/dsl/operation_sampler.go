package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// SetWrapMode sets the wrap mode of the target.
func SetWrapMode(wrapModeProvider Provider[mdl.WrapMode]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			wrapMode, err := wrapModeProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting wrap mode: %w", err)
			}

			wrappable, ok := target.(mdl.Wrappable)
			if !ok {
				return fmt.Errorf("target %T is not a wrappble", target)
			}
			wrappable.SetWrapMode(wrapMode)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-wrap-mode", wrapModeProvider)
		},
	)
}

// SetFilterMode sets the filter mode of the target.
func SetFilterMode(filterModeProvider Provider[mdl.FilterMode]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			filterMode, err := filterModeProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting filter mode: %w", err)
			}

			filterable, ok := target.(mdl.Filterable)
			if !ok {
				return fmt.Errorf("target %T is not a filterable", target)
			}
			filterable.SetFilterMode(filterMode)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-filter-mode", filterModeProvider)
		},
	)
}

// SetMipmapping sets the mipmapping of the target.
func SetMipmapping(mipmappingProvider Provider[bool]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			mipmapping, err := mipmappingProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting mipmapping: %w", err)
			}

			mipmappable, ok := target.(mdl.Mipmappable)
			if !ok {
				return fmt.Errorf("target %T is not a mipmappable", target)
			}
			mipmappable.SetMipmapping(mipmapping)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-mipmapping", mipmappingProvider)
		},
	)
}
