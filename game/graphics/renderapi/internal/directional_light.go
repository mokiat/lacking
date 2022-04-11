package internal

import (
	"github.com/mokiat/lacking/game/graphics/renderapi/plugin"
	"github.com/mokiat/lacking/render"
)

func NewDirectionalLightPresentation(api render.API, shaderSet plugin.ShaderSet) *LightingPresentation {
	return NewLightingPresentation(api, shaderSet.VertexShader(), shaderSet.FragmentShader())
}
