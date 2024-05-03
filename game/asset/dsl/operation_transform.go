package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
)

// SetTranslation sets the translation of the target.
func SetTranslation(translationProvider Provider[dprec.Vec3]) Operation {
	type translatable interface {
		SetTranslation(dprec.Vec3)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			translation, err := translationProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting translation: %w", err)
			}

			transformable, ok := target.(translatable)
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
	type rotatable interface {
		SetRotation(dprec.Quat)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			rotation, err := rotationProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting rotation: %w", err)
			}

			transformable, ok := target.(rotatable)
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
	type scalable interface {
		SetScale(dprec.Vec3)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			scale, err := scaleProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting scale: %w", err)
			}

			transformable, ok := target.(scalable)
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

// SetWidth sets the width of the target.
func SetWidth(widthProvider Provider[float64]) Operation {
	type widthConfigurable interface {
		SetWidth(float64)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			width, err := widthProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting width: %w", err)
			}

			configurable, ok := target.(widthConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not a box", target)
			}
			configurable.SetWidth(width)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-width", widthProvider)
		},
	)
}

// SetHeight sets the width of the target.
func SetHeight(heightProvider Provider[float64]) Operation {
	type heightConfigurable interface {
		SetHeight(float64)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			height, err := heightProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting height: %w", err)
			}

			configurable, ok := target.(heightConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not a box", target)
			}
			configurable.SetHeight(height)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-height", heightProvider)
		},
	)
}

// SetLength sets the width of the target.
func SetLength(lengthProvider Provider[float64]) Operation {
	type lengthConfigurable interface {
		SetLength(float64)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			length, err := lengthProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting length: %w", err)
			}

			configurable, ok := target.(lengthConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not a box", target)
			}
			configurable.SetLength(length)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-length", lengthProvider)
		},
	)
}
