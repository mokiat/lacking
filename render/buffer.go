package render

type BufferInfo struct {
	Dynamic bool
	Data    []byte
	Size    int
}

type BufferUpdateInfo struct {
	Data   []byte
	Offset int
}

type BufferFetchInfo struct {
	Offset int
	Target []byte
}

type BufferObject interface {
	_isBufferObject() bool // ensures interface uniqueness
}

type Buffer interface {
	BufferObject

	// Deprecated: Use queue commands.
	Fetch(info BufferFetchInfo)

	Release()
}
