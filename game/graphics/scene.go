package graphics

// Scene represents a collection of 3D render entities
// that comprise a single visual scene.
type Scene interface {

	// Sky returns this scene's sky object.
	// You can use the Sky object to control the
	// background appearance.
	Sky() Sky

	// CreateCamera creates a new camera object to be
	// used with this scene.
	CreateCamera() Camera

	// CreateLight creates a new light object to be
	// used within this scene.
	CreateLight() Light

	// CreateMesh creates a new mesh instance from the specified
	// template and places it in the scene.
	CreateMesh(template MeshTemplate) Mesh

	// Render draws this scene to the specified viewport
	// looking through the specified camera.
	Render(viewport Viewport, camera Camera)

	// Delete removes this scene and releases all
	// entities allocated for it.
	Delete()
}
