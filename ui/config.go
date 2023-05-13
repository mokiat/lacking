package ui

import (
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/resource"
)

// NewConfig creates a new Config instance.
func NewConfig(locator resource.ReadLocator, renderAPI render.API, shaders ShaderCollection) *Config {
	return &Config{
		locator:   locator,
		renderAPI: renderAPI,
		shaders:   shaders,
	}
}

// Config holds the configuration options for creating a UI controller.
type Config struct {
	locator   resource.ReadLocator
	renderAPI render.API
	shaders   ShaderCollection
}

// ShaderCollection holds the set of shaders to be used for rendering.
type ShaderCollection struct {
	ShapeShadedSet func() ShaderSet
	ShapeBlankSet  func() ShaderSet
	ContourSet     func() ShaderSet
	TextSet        func() ShaderSet
}

// ShaderSet contains the combination of shaders that make a single shader
// program for rendering.
type ShaderSet struct {
	VertexShader   string
	FragmentShader string
}
