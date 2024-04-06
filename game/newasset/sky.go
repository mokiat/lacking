package asset

// Sky represents the background of the scene.
type Sky struct {

	// NodeIndex is the index of the node that the sky is attached to.
	NodeIndex uint32

	// Textures is a list of textures that will be bound to the material.
	//
	// The textures will be bound in the order they are specified.
	Textures []TextureBinding

	// Properties is a list of properties that will be passed to the shader.
	Properties []PropertyBinding

	// Layers is the list of layers that make up the sky.
	Layers []SkyLayer
}

// SkyLayer represents a single layer of the sky.
type SkyLayer struct {

	// Blending specifies whether blending should be applied to the layer.
	Blending bool

	// ShaderIndex is the index of the shader to be used.
	ShaderIndex uint32
}
