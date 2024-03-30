package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/model"
)

func SetTranslation(translation dprec.Vec3) Operation {
	apply := func(target any) error {
		transformable, ok := target.(model.Translatable)
		if !ok {
			return fmt.Errorf("target %T is not a translatable", target)
		}
		transformable.SetTranslation(translation)
		return nil
	}

	digest := func() ([]byte, error) {
		return digestItems("set-translation", translation)
	}

	return FuncOperation(apply, digest)
}

func SetRotation(rotation dprec.Quat) Operation {
	apply := func(target any) error {
		transformable, ok := target.(model.Rotatable)
		if !ok {
			return fmt.Errorf("target %T is not a rotatable", target)
		}
		transformable.SetRotation(rotation)
		return nil
	}

	digest := func() ([]byte, error) {
		return digestItems("set-rotation", rotation)
	}

	return FuncOperation(apply, digest)
}

func SetScale(scale dprec.Vec3) Operation {
	apply := func(target any) error {
		transformable, ok := target.(model.Scalable)
		if !ok {
			return fmt.Errorf("target %T is not a scalable", target)
		}
		transformable.SetScale(scale)
		return nil
	}

	digest := func() ([]byte, error) {
		return digestItems("set-scale", scale)
	}

	return FuncOperation(apply, digest)
}
