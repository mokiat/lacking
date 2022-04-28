package ui

import "github.com/mokiat/lacking/render"

// NewConfig creates a new Config instance.
func NewConfig(locator ResourceLocator, renderAPI render.API, shaders ShaderCollection) *Config {
	return &Config{
		locator:   locator,
		renderAPI: renderAPI,
		shaders:   shaders,
	}
}

// Config holds the configuration options for creating a UI controller.
type Config struct {
	locator   ResourceLocator
	renderAPI render.API
	shaders   ShaderCollection
}

// ShaderCollection holds the set of shaders to be used for rendering.
type ShaderCollection struct {
	ShapeMaterial      ShaderSet
	ShapeBlankMaterial ShaderSet
	ContourMaterial    ShaderSet
	TextMaterial       ShaderSet
}

// ShaderSet contains the combination of shaders that make a single shader
// program for rendering.
type ShaderSet struct {
	VertexShader   func() string
	FragmentShader func() string
}
