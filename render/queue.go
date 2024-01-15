package render

// Queue is the interface that provides the API with the ability
// to schedule commands to be executed on the GPU.
type Queue interface {

	// Invalidate indicates that the graphics state might have changed
	// from outside this API and any cached state by the API should
	// be discarded.
	//
	// Using this command will force a subsequent draw command to initialize
	// all graphics state (e.g. blend state, depth state, stencil state, etc.)
	// on old implementations.
	Invalidate()

	// WriteBuffer writes the provided data to the specified buffer at the
	// specified offset.
	WriteBuffer(buffer Buffer, offset int, data []byte)
}
