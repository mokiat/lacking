package meshdto

var MeshChunkID = "lacking:mesh"

type MeshChunkHolder struct {
	MeshChunk *MeshChunk `chunk:"lacking:mesh"`
}

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
