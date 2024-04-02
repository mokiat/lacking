package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/model"
)

func CreateSky(name string, operations ...Operation) Provider[model.Node] {
	get := func() (model.Node, error) {
		var sky model.Sky
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

func CreateColorSkyLayer(color dprec.Vec3) Provider[model.SkyLayer] {
	shaderProvider := presetColorSkyShader

	get := func() (model.SkyLayer, error) {
		shader, err := shaderProvider.Get()
		if err != nil {
			return model.SkyLayer{}, fmt.Errorf("error getting preset shader: %w", err)
		}
		var layer model.SkyLayer
		layer.SetBlending(false)
		layer.SetProperty("skyColor", color)
		layer.SetShader(shader)
		return layer, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("color-sky-layer", color, shaderProvider)
	}

	return OnceProvider(FuncProvider(get, digest))
}

var presetColorSkyShader = func() Provider[*model.Shader] {
	get := func() (*model.Shader, error) {
		var shader model.Shader
		shader.SetSourceCode(`
		// lsl shading language

		#uniform {
			skyColor vec3,
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
