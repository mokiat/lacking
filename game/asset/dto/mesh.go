package dto

const MeshChunkID = "lacking:mesh"

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

// Mesh represents an instance of a mesh definition.
type Mesh struct {

	// ID is the unique identifier of the mesh within the file.
	ID uint32

	// NodeID is the ID of the node that is used by this mesh.
	NodeID uint32

	// MeshDefinitionID is the ID of the mesh definition that is used by
	// this mesh.
	MeshDefinitionID uint32

	// ArmatureID is the ID of the armature that is used by this mesh.
	//
	// If the mesh does not use an armature, this value is set to
	// UnspecifiedArmatureID.
	ArmatureID uint32
}
