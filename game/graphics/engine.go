package graphics

// Engine represents an entrypoint to 3D graphics rendering.
type Engine interface {

	// Create initializes this 3D engine.
	Create()

	// CreateScene creates a new 3D Scene. Entities managed
	// within a given scene are isolated within that scene.
	CreateScene() Scene

	// Destroy releases resources allocated by this
	// 3D engine.
	Destroy()
}
