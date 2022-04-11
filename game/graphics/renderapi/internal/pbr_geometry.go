package internal

import (
	"github.com/mokiat/lacking/game/graphics/renderapi/plugin"
	"github.com/mokiat/lacking/render"
)

func NewPBRGeometryPresentation(api render.API, shaderSet plugin.ShaderSet) *GeometryPresentation {
	return NewGeometryPresentation(api, shaderSet.VertexShader(), shaderSet.FragmentShader())
}
