package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/mdl"
)

// CreateAmbientLight creates a new ambient light.
func CreateAmbientLight(opts ...Operation) Provider[*mdl.AmbientLight] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.AmbientLight, error) {
			reflectionTexture, err := defaultCubeTextureProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get reflection texture: %w", err)
			}

			refractionTexture, err := defaultCubeTextureProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get refraction texture: %w", err)
			}

			light := mdl.NewAmbientLight()
			light.SetReflectionTexture(reflectionTexture)
			light.SetRefractionTexture(refractionTexture)
			light.SetCastShadow(false)
			for _, opt := range opts {
				if err := opt.Apply(light); err != nil {
					return nil, err
				}
			}
			return light, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-ambient-light", opts)
		},
	))
}

// CreatePointLight creates a new point light.
func CreatePointLight(opts ...Operation) Provider[*mdl.PointLight] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.PointLight, error) {
			light := mdl.NewPointLight()
			light.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetEmitDistance(10.0)
			light.SetCastShadow(false)
			for _, opt := range opts {
				if err := opt.Apply(light); err != nil {
					return nil, err
				}
			}
			return light, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-point-light", opts)
		},
	))
}

// CreateSpotLight creates a new spot light.
func CreateSpotLight(opts ...Operation) Provider[*mdl.SpotLight] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.SpotLight, error) {
			light := mdl.NewSpotLight()
			light.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetEmitDistance(10.0)
			light.SetEmitAngleOuter(dprec.Degrees(90.0))
			light.SetEmitAngleInner(dprec.Degrees(60.0))
			light.SetCastShadow(false)
			for _, opts := range opts {
				if err := opts.Apply(light); err != nil {
					return nil, err
				}
			}
			return light, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-spot-light", opts)
		},
	))
}

// CreateDirectionalLight creates a new directional light.
func CreateDirectionalLight(opts ...Operation) Provider[*mdl.DirectionalLight] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.DirectionalLight, error) {
			light := mdl.NewDirectionalLight()
			light.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
			light.SetCastShadow(false)
			for _, opts := range opts {
				if err := opts.Apply(light); err != nil {
					return nil, err
				}
			}
			return light, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-directional-light", opts)
		},
	))
}
