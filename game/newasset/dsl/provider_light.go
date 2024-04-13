package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

// CreateAmbientLight creates a new ambient light.
func CreateAmbientLight(name string, opts ...Operation) Provider[mdl.Node] {
	return OnceProvider(FuncProvider(
		// get function
		func() (mdl.Node, error) {
			reflectionTexture, err := defaultCubeTextureProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get reflection texture: %w", err)
			}

			refractionTexture, err := defaultCubeTextureProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get refraction texture: %w", err)
			}

			var light mdl.AmbientLight
			light.SetName(name)
			light.SetTranslation(dprec.ZeroVec3())
			light.SetRotation(dprec.IdentityQuat())
			light.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetReflectionTexture(reflectionTexture)
			light.SetRefractionTexture(refractionTexture)
			light.SetCastShadow(false)
			for _, opt := range opts {
				if err := opt.Apply(&light); err != nil {
					return nil, err
				}
			}
			return &light, nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("ambient-light", name, opts)
		},
	))
}

// CreatePointLight creates a new point light.
func CreatePointLight(name string, opts ...Operation) Provider[mdl.Node] {
	return OnceProvider(FuncProvider(
		// get function
		func() (mdl.Node, error) {
			var light mdl.PointLight
			light.SetName(name)
			light.SetTranslation(dprec.ZeroVec3())
			light.SetRotation(dprec.IdentityQuat())
			light.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetEmitDistance(10.0)
			light.SetCastShadow(false)
			for _, opt := range opts {
				if err := opt.Apply(&light); err != nil {
					return nil, err
				}
			}
			return &light, nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("point-light", name, opts)
		},
	))
}

// CreateSpotLight creates a new spot light.
func CreateSpotLight(name string, opts ...Operation) Provider[mdl.Node] {
	return OnceProvider(FuncProvider(
		// get function
		func() (mdl.Node, error) {
			var light mdl.SpotLight
			light.SetName(name)
			light.SetTranslation(dprec.ZeroVec3())
			light.SetRotation(dprec.IdentityQuat())
			light.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetEmitDistance(10.0)
			light.SetEmitAngleOuter(dprec.Degrees(90.0))
			light.SetEmitAngleInner(dprec.Degrees(60.0))
			light.SetCastShadow(false)
			for _, opts := range opts {
				if err := opts.Apply(&light); err != nil {
					return nil, err
				}
			}
			return &light, nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("spot-light", name, opts)
		},
	))
}
