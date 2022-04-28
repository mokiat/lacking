package ui

import "github.com/mokiat/lacking/render"

func NewConfig(locator ResourceLocator, renderAPI render.API, shaders ShaderCollection) *Config {
	return &Config{
		locator:   locator,
		renderAPI: renderAPI,
		shaders:   shaders,
	}
}

type Config struct {
	locator   ResourceLocator
	renderAPI render.API
	shaders   ShaderCollection
}

type ShaderCollection struct {
	ShapeMaterial      ShaderSet
	ShapeBlankMaterial ShaderSet
	ContourMaterial    ShaderSet
	TextMaterial       ShaderSet
}

type ShaderSet struct {
	VertexShader   func() string
	FragmentShader func() string
}
