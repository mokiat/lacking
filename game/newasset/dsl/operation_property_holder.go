package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

// SetProperty sets the specified property on the target property holder.
func SetProperty[T any](name string, valueProvider Provider[T]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			value, err := valueProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting value: %w", err)
			}

			container, ok := target.(mdl.PropertyHolder)
			if !ok {
				return fmt.Errorf("target %T is not a property holder", target)
			}
			container.SetProperty(name, value)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-property", name, valueProvider)
		},
	)
}
