package ui

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
