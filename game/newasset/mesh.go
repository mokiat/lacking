package asset

const (
	UnspecifiedArmatureIndex = int32(-1)
)

// MaterialBinding represents the binding of a material to a geometry fragment.
type MaterialBinding struct {

	// MaterialIndex is the index of the material that is bound to the fragment.
	MaterialIndex uint32

	// FragmentIndex is the index of the fragment that is bound to the material.
	FragmentIndex uint32
}

// MeshDefinition represents the definition of a mesh. It extends the Geometry
// definition by adding material bindings.
type MeshDefinition struct {

	// Name is the name of the mesh definition.
	Name string

	// GeometryIndex is the index of the geometry that is used by this mesh.
	GeometryIndex uint32

	// MaterialBindings is the collection of material bindings that are used by
	// this mesh.
	MaterialBindings []MaterialBinding
}

// Mesh represents an instance of a mesh definition.
type Mesh struct {

	// MeshDefinitionIndex is the index of the mesh definition that is used by
	// this mesh.
	MeshDefinitionIndex uint32

	// ArmatureIndex is the index of the armature that is used by this mesh.
	//
	// If the mesh does not use an armature, this value is set to
	// UnspecifiedArmatureIndex.
	ArmatureIndex int32

	// NodeIndex is the index of the node that is used by this mesh.
	NodeIndex uint32
}
