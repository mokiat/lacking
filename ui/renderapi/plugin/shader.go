package plugin

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
