package asset

// Scene represents a virtual world that is composed of various visual
// and logical elements.
type Scene struct {

	// Nodes is the collection of nodes that are part of the scene.
	Nodes []Node

	// Animations is the collection of animations that are part of the scene.
	Armatures []Armature

	// Geometries is the collection of geometries that are part of the scene.
	Geometries []Geometry

	// MeshDefinitions is the collection of mesh definitions that are part of
	// the scene.
	MeshDefinitions []MeshDefinition

	// Meshes is the collection of mesh instances that are part of the scene.
	Meshes []Mesh

	// AmbientLights is the collection of ambient lights that are part of the
	// scene.
	AmbientLights []AmbientLight

	// PointLights is the collection of point lights that are part of the scene.
	PointLights []PointLight

	// SpotLights is the collection of spot lights that are part of the scene.
	SpotLights []SpotLight

	// DirectionalLights is the collection of directional lights that are part
	// of the scene.
	DirectionalLights []DirectionalLight

	// SceneDefinitions is the collection of external scene definitions that
	// are used by the scene.
	SceneDefinitions []string

	// SceneInstances is the instantiation of external scene definitions within
	// this scene.
	SceneInstances []SceneInstance
}

// SceneInstance represents the instantiation of an external scene definition
// within a scene.
type SceneInstance struct {

	// SceneDefinitionIndex is the index of the scene definition that is used
	// by this scene instance.
	SceneDefinitionIndex uint32

	// NodeIndex is the index of the node that is used by this scene instance.
	NodeIndex uint32
}
