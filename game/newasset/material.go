package asset

// TextureBinding represents a binding of a texture to a shader.
type TextureBinding struct {

	// TextureIndex is the index of the texture to be bound.
	TextureIndex uint32

	// Wrapping specifies the texture wrapping mode.
	Wrapping WrapMode

	// Filtering specifies the texture filtering mode.
	Filtering FilterMode

	// Mipmapping specifies whether mipmapping should be applied.
	Mipmapping bool
}

// GeometryPass represents a pass that is applied to the geometry of a mesh.
type GeometryPass struct {

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

	// Textures is a list of textures that will be bound to the material.
	//
	// The textures will be bound in the order they are specified.
	Textures []TextureBinding

	// MaterialDataStd140 is the material data that will be passed to the
	// shader. It must be in std140 layout.
	MaterialDataStd140 []byte

	// ShaderIndex is the index of the shader to be used.
	ShaderIndex uint32
}

// ShadowPass represents a pass that is applied to form the shadow of a mesh.
type ShadowPass struct {

	// Layer controls the render ordering of this pass. Lower values will be
	// rendered first. Having too many layers can affect performance.
	Layer int32

	// Culling specifies the culling mode.
	Culling CullMode

	// FrontFace specifies the front face orientation.
	FrontFace FaceOrientation

	// Textures is a list of textures that will be bound to the material.
	//
	// The textures will be bound in the order they are specified.
	Textures []TextureBinding

	// MaterialDataStd140 is the material data that will be passed to the
	// shader. It must be in std140 layout.
	MaterialDataStd140 []byte

	// ShaderIndex is the index of the shader to be used.
	ShaderIndex uint32
}

// ForwardPass represents a pass that is applied to the mesh during the forward
// rendering phase.
type ForwardPass struct {

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

	// Textures is a list of textures that will be bound to the material.
	//
	// The textures will be bound in the order they are specified.
	Textures []TextureBinding

	// MaterialDataStd140 is the material data that will be passed to the
	// shader. It must be in std140 layout.
	MaterialDataStd140 []byte

	// ShaderIndex is the index of the shader to be used.
	ShaderIndex uint32
}

// Material represents a material that can be applied to a mesh.
type Material struct {

	// Name is the name of the material.
	Name string

	// GeometryPasses specifies a list of geometry passes to be applied.
	GeometryPasses []GeometryPass

	// ShadowPasses specifies a list of shadow passes to be applied.
	ShadowPasses []ShadowPass

	// ForwardPasses specifies a list of forward passes to be applied.
	ForwardPasses []ForwardPass
}
