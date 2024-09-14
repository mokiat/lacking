package render

// BufferMarker marks a type as being a Buffer.
type BufferMarker interface {
	_isBufferType()
}

// Buffer is used to store data that can be used by the GPU.
type Buffer interface {
	BufferMarker
	Resource

	// Label returns a human-readable name for the Buffer.
	Label() string
}

// BufferInfo describes the information needed to create a new Buffer.
type BufferInfo struct {

	// Label specifies a human-readable label for the Buffer. Intended for
	// debugging and logging purposes only.
	Label string

	// Size specifies the size of the buffer in bytes.
	Size uint32

	// Data specifies the initial data that should be stored in the buffer.
	// This can be nil if no initial data is needed.
	Data []byte

	// Dynamic specifies whether the buffer is intended to be updated.
	Dynamic bool
}
