package dto

// Model represents a virtual world that is composed of various visual
// and logical elements.
type Model struct {
	HierarchyChunkHolder
	AnimationChunkHolder
	ShadingChunkHolder
	LightingChunkHolder
	MeshChunkHolder
	PhysicsChunkHolder
	CameraChunkHolder
	BackgroundChunkHolder
}
