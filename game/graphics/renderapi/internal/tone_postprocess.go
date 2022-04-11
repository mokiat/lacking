package internal

import (
	"github.com/mokiat/lacking/game/graphics/renderapi/plugin"
	"github.com/mokiat/lacking/render"
)

func NewTonePostprocessingPresentation(api render.API, shaderSet plugin.ShaderSet) *PostprocessingPresentation {
	return NewPostprocessingPresentation(api, shaderSet.VertexShader(), shaderSet.FragmentShader())
}
