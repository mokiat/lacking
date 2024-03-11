package render

// BufferMarker marks a type as being a Buffer.
type BufferMarker interface {
	_isBufferType()
}

// Buffer is used to store data that can be used by the GPU.
type Buffer interface {
	BufferMarker
	Resource
}

// BufferInfo describes the information needed to create a new Buffer.
type BufferInfo struct {
	Dynamic bool
	Data    []byte
	Size    int
}
