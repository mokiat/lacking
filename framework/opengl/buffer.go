package opengl

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewBuffer() *Buffer {
	return &Buffer{}
}

type Buffer struct {
	id      uint32
	dynamic bool
}

func (b *Buffer) ID() uint32 {
	return b.id
}

func (b *Buffer) Allocate(info BufferAllocateInfo) {
	if b.id != 0 {
		panic(fmt.Errorf("buffer already allocated"))
	}
	gl.CreateBuffers(1, &b.id)
	if b.id == 0 {
		panic(fmt.Errorf("failed to allocate buffer"))
	}
	gl.NamedBufferStorage(b.id, len(info.Data), gl.Ptr(info.Data), info.glFlags())
	runtime.KeepAlive(info.Data)
	b.dynamic = info.Dynamic
}

func (b *Buffer) Update(info BufferUpdateInfo) {
	if b.id == 0 {
		panic(fmt.Errorf("trying to update a released buffer"))
	}
	if !b.dynamic {
		panic(fmt.Errorf("trying to update a static buffer"))
	}
	gl.NamedBufferSubData(b.id, info.OffsetBytes, len(info.Data), gl.Ptr(info.Data))
	runtime.KeepAlive(info.Data)
}

func (b *Buffer) Release() {
	if b.id == 0 {
		panic(fmt.Errorf("buffer already released"))
	}
	gl.DeleteBuffers(1, &b.id)
	b.id = 0
	b.dynamic = false
}

type BufferAllocateInfo struct {
	Dynamic bool
	Data    []byte
}

func (i BufferAllocateInfo) glFlags() uint32 {
	var flags uint32
	if i.Dynamic {
		flags |= gl.DYNAMIC_STORAGE_BIT
	}
	return flags
}

type BufferUpdateInfo struct {
	Data        []byte
	OffsetBytes int
}
