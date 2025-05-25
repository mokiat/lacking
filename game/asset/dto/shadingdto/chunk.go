package shadingdto

type ShadingChunkHolder struct {
	ShadingChunk *ShadingChunk `chunk:"lacking:shading"`
}

type ShadingChunk struct {
	// Shaders is the collection of custom shaders that are are to be used.
	Shaders []Shader

	// Textures is the collection of textures that are part of the scene.
	Textures []Texture

	// Materials is the collection of materials that are part of the scene.
	Materials []Material
}
