package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

// SetTranslation sets the translation of the target.
func SetTranslation(translationProvider Provider[dprec.Vec3]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			translation, err := translationProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting translation: %w", err)
			}

			transformable, ok := target.(mdl.Translatable)
			if !ok {
				return fmt.Errorf("target %T is not a translatable", target)
			}
			transformable.SetTranslation(translation)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-translation", translationProvider)
		},
	)
}

// SetRotation sets the rotation of the target.
func SetRotation(rotationProvider Provider[dprec.Quat]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			rotation, err := rotationProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting rotation: %w", err)
			}

			transformable, ok := target.(mdl.Rotatable)
			if !ok {
				return fmt.Errorf("target %T is not a rotatable", target)
			}
			transformable.SetRotation(rotation)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-rotation", rotationProvider)
		},
	)
}

// SetScale sets the scale of the target.
func SetScale(scaleProvider Provider[dprec.Vec3]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			scale, err := scaleProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting scale: %w", err)
			}

			transformable, ok := target.(mdl.Scalable)
			if !ok {
				return fmt.Errorf("target %T is not a scalable", target)
			}
			transformable.SetScale(scale)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-scale", scaleProvider)
		},
	)
}
