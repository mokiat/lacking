package meshdto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var meshChunkID = gog.Must(uuid.Parse("0332375f-64de-4dab-ad28-ae556781987b"))

type MeshChunk struct {
	// Armatures is the collection of armatures that are part of the scene.
	Armatures []Armature

	// Geometries is the collection of geometries that are part of the scene.
	Geometries []Geometry

	// MeshDefinitions is the collection of mesh definitions that are part of
	// the scene.
	MeshDefinitions []MeshDefinition

	// Meshes is the collection of mesh instances that are part of the scene.
	Meshes []Mesh
}

func (c MeshChunk) ChunkID() uuid.UUID {
	return meshChunkID
}
