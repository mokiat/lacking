package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func SetWrapMode(wrapMode mdl.WrapMode) Operation {
	apply := func(target any) error {
		wrappable, ok := target.(mdl.Wrappable)
		if !ok {
			return fmt.Errorf("target %T is not a wrappble", target)
		}
		wrappable.SetWrapMode(wrapMode)
		return nil
	}

	digest := func() ([]byte, error) {
		return CreateDigest("set-wrap-mode", uint8(wrapMode))
	}

	return FuncOperation(apply, digest)
}

func SetFilterMode(filterMode mdl.FilterMode) Operation {
	apply := func(target any) error {
		filterable, ok := target.(mdl.Filterable)
		if !ok {
			return fmt.Errorf("target %T is not a filterable", target)
		}
		filterable.SetFilterMode(filterMode)
		return nil
	}

	digest := func() ([]byte, error) {
		return CreateDigest("set-filter-mode", uint8(filterMode))
	}

	return FuncOperation(apply, digest)
}

func SetMipmapping(mipmapping bool) Operation {
	apply := func(target any) error {
		mipmappable, ok := target.(mdl.Mipmappable)
		if !ok {
			return fmt.Errorf("target %T is not a mipmappable", target)
		}
		mipmappable.SetMipmapping(mipmapping)
		return nil
	}

	digest := func() ([]byte, error) {
		return CreateDigest("set-mipmapping", mipmapping)
	}

	return FuncOperation(apply, digest)
}
