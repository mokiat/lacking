package opengl

import (
	"fmt"

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

func (b *Buffer) Allocate(info BufferAllocateInfo) error {
	gl.CreateBuffers(1, &b.id)
	if b.id == 0 {
		return fmt.Errorf("failed to allocate buffer")
	}
	gl.NamedBufferStorage(b.id, len(info.Data), gl.Ptr(info.Data), info.glFlags())
	b.dynamic = info.Dynamic
	return nil
}

func (b *Buffer) Update(info BufferUpdateInfo) error {
	if !b.dynamic {
		return fmt.Errorf("trying to update a static index buffer")
	}
	gl.NamedBufferSubData(b.id, info.OffsetBytes, len(info.Data), gl.Ptr(info.Data))
	return nil
}

func (b *Buffer) Release() error {
	gl.DeleteBuffers(1, &b.id)
	b.id = 0
	b.dynamic = false
	return nil
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
