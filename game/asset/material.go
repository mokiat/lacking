package asset

// TextureBinding represents a binding of a texture to a shader.
type TextureBinding struct {

	// BindingName is the name of the binding in the shader.
	BindingName string

	// TextureIndex is the index of the texture to be bound.
	TextureIndex uint32

	// Wrapping specifies the texture wrapping mode.
	Wrapping WrapMode

	// Filtering specifies the texture filtering mode.
	Filtering FilterMode

	// Mipmapping specifies whether mipmapping should be applied.
	Mipmapping bool
}

// PropertyBinding represents a binding of a uniform property to a shader.
type PropertyBinding struct {

	// BindingName is the name of the binding in the shader.
	BindingName string

	// Data is the data to be bound.
	Data []byte
}

// MaterialPass represents a pass that is applied during material rendering.
type MaterialPass struct {

	// Layer controls the render ordering of this pass. Lower values will be
	// rendered first. Having too many layers can affect performance.
	Layer int32

	// Culling specifies the culling mode.
	Culling CullMode

	// FrontFace specifies the front face orientation.
	FrontFace FaceOrientation

	// DepthTest specifies whether depth testing should be enabled.
	DepthTest bool

	// DepthWrite specifies whether depth writing should be enabled.
	DepthWrite bool

	// DepthComparison specifies the depth comparison function.
	DepthComparison Comparison

	// Blending specifies whether blending should be enabled.
	Blending bool

	// ShaderIndex is the index of the shader to be used.
	ShaderIndex uint32
}

// Material represents a material that can be applied to a mesh.
type Material struct {

	// Name is the name of the material.
	Name string

	// Textures is a list of textures that will be bound to the material.
	Textures []TextureBinding

	// Properties is a list of properties that will be passed to the shader.
	Properties []PropertyBinding

	// GeometryPasses specifies a list of geometry passes to be applied.
	GeometryPasses []MaterialPass

	// ShadowPasses specifies a list of shadow passes to be applied.
	ShadowPasses []MaterialPass

	// ForwardPasses specifies a list of forward passes to be applied.
	ForwardPasses []MaterialPass

	// SkyPasses specifies a list of sky passes to be applied,
	// applicable only to sky materials.
	SkyPasses []MaterialPass

	// PostprocessingPasses specifies a list of postprocessing passes to
	// be applied. Applicable only to postprocessing materials.
	PostprocessingPasses []MaterialPass
}
