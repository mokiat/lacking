package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/mdl"
)

// CreateSky creates a new sky with the provided name and operations.
func CreateSky(name string, operations ...Operation) Provider[mdl.Node] {
	return OnceProvider(FuncProvider(
		// get function
		func() (mdl.Node, error) {
			var sky mdl.Sky
			sky.SetName(name)
			sky.SetTranslation(dprec.ZeroVec3())
			sky.SetRotation(dprec.IdentityQuat())
			sky.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
			for _, operation := range operations {
				if err := operation.Apply(&sky); err != nil {
					return nil, err
				}
			}
			return &sky, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-sky", name, operations)
		},
	))
}

// CreateColorSkyLayer creates a new color sky layer.
func CreateColorSkyLayer() Provider[mdl.SkyLayer] {
	return OnceProvider(FuncProvider(
		// get function
		func() (mdl.SkyLayer, error) {
			shader, err := defaultColorSkyShader.Get()
			if err != nil {
				return mdl.SkyLayer{}, fmt.Errorf("error getting preset shader: %w", err)
			}
			var layer mdl.SkyLayer
			layer.SetBlending(false)
			layer.SetShader(shader)
			return layer, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-color-sky-layer")
		},
	))
}

// CreateTextureSkyLayer creates a new texture sky layer.
func CreateTextureSkyLayer() Provider[mdl.SkyLayer] {
	return OnceProvider(FuncProvider(
		// get function
		func() (mdl.SkyLayer, error) {
			shader, err := defaultTextureSkyShader.Get()
			if err != nil {
				return mdl.SkyLayer{}, fmt.Errorf("error getting preset shader: %w", err)
			}
			var layer mdl.SkyLayer
			layer.SetBlending(false)
			layer.SetShader(shader)
			return layer, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-texture-sky-layer")
		},
	))
}

var defaultColorSkyShader = func() Provider[*mdl.Shader] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Shader, error) {
			var shader mdl.Shader
			shader.SetSourceCode(`
				uniforms {
					skyColor vec4,
				}
		
				func #fragment() {
					#color = skyColor
				}
			`)
			return &shader, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("default-color-sky-shader")
		},
	))
}()

var defaultTextureSkyShader = func() Provider[*mdl.Shader] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Shader, error) {
			var shader mdl.Shader
			shader.SetSourceCode(`
				textures {
					skyColorSampler samplerCube,
				}
		
				func #fragment() {
					#color = sample(skyColorSampler, #direction)
				}
			`)
			return &shader, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("default-texture-sky-shader")
		},
	))
}()
