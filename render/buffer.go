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

type BufferObject interface {
	_isBufferObject() bool // ensures interface uniqueness
}

type Buffer interface {
	BufferObject
	Update(info BufferUpdateInfo)
	Release()
}
