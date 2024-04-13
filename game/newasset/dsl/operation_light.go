package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

// SetEmitColor configures the emit color of the target.
func SetEmitColor(colorProvider Provider[dprec.Vec3]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			color, err := colorProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting emit color: %w", err)
			}

			emitter, ok := target.(mdl.ColorEmitter)
			if !ok {
				return fmt.Errorf("target %T is not a color emitter", target)
			}
			emitter.SetEmitColor(color)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("set-emit-color", colorProvider)
		},
	)
}

// SetEmitDistance configures the emit distance of the target.
func SetEmitDistance(distanceProvider Provider[float64]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			distance, err := distanceProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting emit distance: %w", err)
			}

			emitter, ok := target.(mdl.DistanceEmitter)
			if !ok {
				return fmt.Errorf("target %T is not a distance emitter", target)
			}
			emitter.SetEmitDistance(distance)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("set-emit-distance", distanceProvider)
		},
	)
}

// SetEmitAngleOuter configures the outer emit angle of the target.
func SetEmitAngleOuter(angleProvider Provider[dprec.Angle]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			angle, err := angleProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting outer emit angle: %w", err)
			}

			emitter, ok := target.(mdl.ConeEmitter)
			if !ok {
				return fmt.Errorf("target %T is not a cone emitter", target)
			}
			emitter.SetEmitAngleOuter(angle)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("set-emit-angle-outer", angleProvider)
		},
	)
}

// SetEmitAngleInner configures the inner emit angle of the target.
func SetEmitAngleInner(angleProvider Provider[dprec.Angle]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			angle, err := angleProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting inner emit angle: %w", err)
			}

			emitter, ok := target.(mdl.ConeEmitter)
			if !ok {
				return fmt.Errorf("target %T is not a cone emitter", target)
			}
			emitter.SetEmitAngleInner(angle)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("set-emit-angle-inner", angleProvider)
		},
	)
}

// SetReflectionTexture configures the reflection texture of the target.
func SetReflectionTexture(textureProvider Provider[*mdl.Texture]) Operation {
	type reflectionTextureConfigurable interface {
		SetReflectionTexture(*mdl.Texture)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			texture, err := textureProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting reflection texture: %w", err)
			}

			configurable, ok := target.(reflectionTextureConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not configurable with a reflection texture", target)
			}
			configurable.SetReflectionTexture(texture)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("set-reflection-texture", textureProvider)
		},
	)
}

// SetRefractionTexture configures the refraction texture of the target.
func SetRefractionTexture(textureProvider Provider[*mdl.Texture]) Operation {
	type refractionTextureConfigurable interface {
		SetRefractionTexture(*mdl.Texture)
	}

	return FuncOperation(
		// apply function
		func(target any) error {
			texture, err := textureProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting refraction texture: %w", err)
			}

			configurable, ok := target.(refractionTextureConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not configurable with a refraction texture", target)
			}
			configurable.SetRefractionTexture(texture)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("set-refraction-texture", textureProvider)
		},
	)
}
