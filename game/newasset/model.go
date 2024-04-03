package asset

// Model represents a virtual world that is composed of various visual
// and logical elements.
type Model struct {

	// Nodes is the collection of nodes that are part of the scene.
	Nodes []Node

	// Cameras is the collection of cameras that are part of the scene.
	Cameras []Camera

	// Shaders is the collection of shaders that are part of the scene.
	Shaders []Shader

	// Materials is the collection of materials that are part of the scene.
	Materials []Material

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

	// Skies is the collection of skies that are part of the scene.
	Skies []Sky

	// ModelDefinitions is the collection of external scene definitions that
	// are used by the scene.
	ModelDefinitions []string

	// ModelInstances is the instantiation of external scene definitions within
	// this scene.
	ModelInstances []ModelInstance
}

// ModelInstance represents the instantiation of an external scene definition
// within a scene.
type ModelInstance struct {

	// SceneDefinitionIndex is the index of the scene definition that is used
	// by this scene instance.
	SceneDefinitionIndex uint32

	// NodeIndex is the index of the node that is used by this scene instance.
	NodeIndex uint32
}
