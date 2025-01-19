package asset

// Model represents a virtual world that is composed of various visual
// and logical elements.
type Model struct {

	// Nodes is the collection of nodes that are part of the scene.
	Nodes []Node

	// Animations is the collection of animations that are part of the scene.
	Animations []Animation

	// Armatures is the collection of armatures that are part of the scene.
	Armatures []Armature

	// Cameras is the collection of cameras that are part of the scene.
	Cameras []Camera

	// Shaders is the collection of custom shaders that are are to be used.
	Shaders []Shader

	// Textures is the collection of textures that are part of the scene.
	Textures []Texture

	// Materials is the collection of materials that are part of the scene.
	Materials []Material

	// Geometries is the collection of geometries that are part of the scene.
	Geometries []Geometry

	// MeshDefinitions is the collection of mesh definitions that are part of
	// the scene.
	MeshDefinitions []MeshDefinition

	// Meshes is the collection of mesh instances that are part of the scene.
	Meshes []Mesh

	// BodyMaterials is the collection of body materials that are part of the
	// scene.
	BodyMaterials []BodyMaterial

	// BodyDefinitions is the collection of body definitions that are part of
	// the scene.
	BodyDefinitions []BodyDefinition

	// Bodies is the collection of body instances that are part of the scene.
	Bodies []Body

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

	// Blobls is the collection of binary blobs that are part of the scene.
	Blobs []Blob
}

// ModelInstance represents the instantiation of an external scene definition
// within a scene.
type ModelInstance struct {

	// ModelDefinitionIndex is the index of the scene definition that is used
	// by this scene instance.
	ModelDefinitionIndex uint32

	// NodeIndex is the index of the node that is used by this scene instance.
	NodeIndex uint32
}
