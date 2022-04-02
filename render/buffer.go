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

type Buffer interface {
	Update(info BufferUpdateInfo)
	Release()
}
