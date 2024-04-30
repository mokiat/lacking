package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/asset/mdl"
)

// CreateSky creates a new sky with the provided name and operations.
func CreateSky(name string, materialProvider Provider[*mdl.Material], opts ...Operation) Provider[mdl.Node] {
	return OnceProvider(FuncProvider(
		// get function
		func() (mdl.Node, error) {
			material, err := materialProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting material: %w", err)
			}

			var sky mdl.Sky
			sky.SetName(name)
			sky.SetMaterial(material)
			sky.SetTranslation(dprec.ZeroVec3())
			sky.SetRotation(dprec.IdentityQuat())
			sky.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
			for _, opt := range opts {
				if err := opt.Apply(&sky); err != nil {
					return nil, err
				}
			}
			return &sky, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-sky", name, opts)
		},
	))
}

// CreateColorSkyMaterial creates a new color sky material.
func CreateColorSkyMaterial(colorProvider Provider[dprec.Vec4]) Provider[*mdl.Material] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Material, error) {
			shader, err := defaultColorSkyShader.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting preset shader: %w", err)
			}

			color, err := colorProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting color: %w", err)
			}

			var pass mdl.MaterialPass
			pass.SetLayer(0)
			pass.SetCulling(mdl.CullModeNone)
			pass.SetFrontFace(mdl.FaceOrientationCW)
			pass.SetDepthTest(true)
			pass.SetDepthWrite(false)
			pass.SetDepthComparison(mdl.ComparisonLess)
			pass.SetBlending(false)
			pass.SetShader(shader)

			var material mdl.Material
			material.SetName("ColorSkyMaterial")
			material.AddSkyPass(&pass)
			material.SetProperty("skyColor", dtos.Vec4(color))
			return &material, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-color-sky-material", colorProvider)
		},
	))
}

// CreateTextureSkyMaterial creates a new texture sky material.
func CreateTextureSkyMaterial(samplerProvider Provider[*mdl.Sampler]) Provider[*mdl.Material] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Material, error) {
			shader, err := defaultTextureSkyShader.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting preset shader: %w", err)
			}

			sampler, err := samplerProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting sampler: %w", err)
			}

			var pass mdl.MaterialPass
			pass.SetLayer(0)
			pass.SetCulling(mdl.CullModeNone)
			pass.SetFrontFace(mdl.FaceOrientationCW)
			pass.SetDepthTest(true)
			pass.SetDepthWrite(false)
			pass.SetDepthComparison(mdl.ComparisonLess)
			pass.SetBlending(false)
			pass.SetShader(shader)

			var material mdl.Material
			material.SetName("ColorSkyMaterial")
			material.AddSkyPass(&pass)
			material.SetSampler("skyColorSampler", sampler)
			return &material, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-texture-sky-material", samplerProvider)
		},
	))
}

var defaultColorSkyShader = func() Provider[*mdl.Shader] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Shader, error) {
			var shader mdl.Shader
			shader.SetShaderType(mdl.ShaderTypeSky)
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
			shader.SetShaderType(mdl.ShaderTypeSky)
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
