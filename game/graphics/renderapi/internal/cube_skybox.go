package internal

import (
	"github.com/mokiat/lacking/game/graphics/renderapi/plugin"
	"github.com/mokiat/lacking/render"
)

func NewCubeSkyboxPresentation(api render.API, shaderSet plugin.ShaderSet) *SkyboxPresentation {
	return NewSkyboxPresentation(api, shaderSet.VertexShader(), shaderSet.FragmentShader())
}

func NewColorSkyboxPresentation(api render.API, shaderSet plugin.ShaderSet) *SkyboxPresentation {
	return NewSkyboxPresentation(api, shaderSet.VertexShader(), shaderSet.FragmentShader())
}
