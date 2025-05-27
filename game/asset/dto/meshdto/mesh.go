package meshdto

const (
	UnspecifiedArmatureIndex = int32(-1)

	UnspecifiedArmatureID = uint32(0xFFFFFFFF)
)

// MaterialBinding represents the binding of a material to a geometry fragment.
type MaterialBinding struct {

	// FragmentIndex is the index of the fragment that is bound to the material.
	FragmentIndex uint32

	// MaterialID is the ID of the material that is bound to the fragment.
	MaterialID uint32
}

// MeshDefinition represents the definition of a mesh. It extends the Geometry
// definition by adding material bindings.
type MeshDefinition struct {

	// ID is the unique identifier of the mesh definition within the file.
	ID uint32

	// GeometryID is the ID of the geometry that is used by this mesh.
	GeometryID uint32

	// MaterialBindings is the collection of material bindings that are used by
	// this mesh.
	MaterialBindings []MaterialBinding
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
