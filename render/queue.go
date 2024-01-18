package render

// QueueMarker marks a type as being a Queue.
type QueueMarker interface {
	_isQueueType()
}

// Queue is the interface that provides the API with the ability
// to schedule commands to be executed on the GPU.
type Queue interface {
	QueueMarker

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

	// ReadBuffer reads data from the specified buffer at the specified offset
	// into the provided target slice.
	ReadBuffer(buffer Buffer, offset int, target []byte)

	// Submit schedules the provided command buffer to be executed on the GPU.
	//
	// Once submitted the command buffer is reset and can be reused.
	Submit(commands CommandBuffer)

	// TrackSubmittedWorkDone creates a fence that will be signaled once all
	// work submitted to the queue before this call has been completed.
	TrackSubmittedWorkDone() Fence
}
