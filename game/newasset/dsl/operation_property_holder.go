package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func SetProperty(name string, value any) Operation {
	apply := func(target any) error {
		container, ok := target.(mdl.PropertyHolder)
		if !ok {
			return fmt.Errorf("target %T is not a property holder", target)
		}
		container.SetProperty(name, value)
		return nil
	}

	digest := func() ([]byte, error) {
		return CreateDigest("set-property", name, value)
	}

	return FuncOperation(apply, digest)
}
