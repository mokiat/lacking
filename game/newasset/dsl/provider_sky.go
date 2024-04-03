package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

func CreateSky(name string, operations ...Operation) Provider[mdl.Node] {
	get := func() (mdl.Node, error) {
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
	}

	digest := func() ([]byte, error) {
		return digestItems("sky", name, operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}

func CreateColorSkyLayer(color dprec.Vec3) Provider[mdl.SkyLayer] {
	shaderProvider := presetColorSkyShader

	get := func() (mdl.SkyLayer, error) {
		shader, err := shaderProvider.Get()
		if err != nil {
			return mdl.SkyLayer{}, fmt.Errorf("error getting preset shader: %w", err)
		}
		var layer mdl.SkyLayer
		layer.SetBlending(false)
		layer.SetProperty("skyColor",
			sprec.NewVec4(float32(color.X), float32(color.Y), float32(color.Z), 1.0),
		)
		layer.SetShader(shader)
		return layer, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("color-sky-layer", color, shaderProvider)
	}

	return OnceProvider(FuncProvider(get, digest))
}

var presetColorSkyShader = func() Provider[*mdl.Shader] {
	get := func() (*mdl.Shader, error) {
		var shader mdl.Shader
		shader.SetSourceCode(`
		// lsl shading language

		#uniform {
			skyColor vec4,
		}

		func #fragment() {
			#color = skyColor
		}
		`)
		return &shader, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("color-sky-shader")
	}

	return OnceProvider(FuncProvider(get, digest))
}()
