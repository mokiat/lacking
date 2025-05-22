package shadingdto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var shadingChunkID = gog.Must(uuid.Parse("4a344e51-4858-4cb8-b907-98b3d799686c"))

type ShadingChunk struct {
	// Shaders is the collection of custom shaders that are are to be used.
	Shaders []Shader

	// Textures is the collection of textures that are part of the scene.
	Textures []Texture

	// Materials is the collection of materials that are part of the scene.
	Materials []Material
}

func (c ShadingChunk) ChunkID() uuid.UUID {
	return shadingChunkID
}
