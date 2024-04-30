package dsl

import (
	"fmt"
)

// BindProperty sets the specified property on the target property holder.
func BindProperty[T any](name string, valueProvider Provider[T]) Operation {
	type propertyHolder interface {
		SetProperty(name string, value any)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			value, err := valueProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting value: %w", err)
			}

			container, ok := target.(propertyHolder)
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
