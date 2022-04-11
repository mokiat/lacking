package internal

import (
	"github.com/mokiat/lacking/game/graphics/renderapi/plugin"
	"github.com/mokiat/lacking/render"
)

func NewExposurePresentation(api render.API, set plugin.ShaderSet) *LightingPresentation {
	return NewLightingPresentation(api, set.VertexShader(), set.FragmentShader())
}
