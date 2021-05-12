package opengl

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewVertexArray() *VertexArray {
	return &VertexArray{}
}

type VertexArray struct {
	id uint32
}

func (a *VertexArray) ID() uint32 {
	return a.id
}

func (a *VertexArray) Allocate(info VertexArrayAllocateInfo) {
	if a.id != 0 {
		panic(fmt.Errorf("vertex array already allocated"))
	}
	gl.CreateVertexArrays(1, &a.id)
	if a.id == 0 {
		panic(fmt.Errorf("failed to allocate vertex array"))
	}
	for index, binding := range info.BufferBindings {
		gl.VertexArrayVertexBuffer(a.id, uint32(index), binding.VertexBuffer.ID(), binding.OffsetBytes, binding.StrideBytes)
	}
	for _, attribute := range info.Attributes {
		gl.EnableVertexArrayAttrib(a.id, attribute.Index)
		gl.VertexArrayAttribFormat(a.id, attribute.Index, attribute.ComponentCount, attribute.ComponentType, attribute.Normalized, attribute.OffsetBytes)
		gl.VertexArrayAttribBinding(a.id, attribute.Index, attribute.BufferBinding)
	}
	if info.IndexBuffer != nil {
		gl.VertexArrayElementBuffer(a.id, info.IndexBuffer.ID())
	}
}

func (a *VertexArray) Release() {
	if a.id == 0 {
		panic(fmt.Errorf("buffer already released"))
	}
	gl.DeleteVertexArrays(1, &a.id)
	a.id = 0
}

type VertexArrayAllocateInfo struct {
	BufferBindings []VertexArrayBufferBinding
	Attributes     []VertexArrayAttribute
	IndexBuffer    *Buffer
}

func NewVertexArrayBufferBinding(buffer *Buffer, offsetBytes int, strideBytes int32) VertexArrayBufferBinding {
	return VertexArrayBufferBinding{
		VertexBuffer: buffer,
		OffsetBytes:  offsetBytes,
		StrideBytes:  strideBytes,
	}
}

type VertexArrayBufferBinding struct {
	VertexBuffer *Buffer
	OffsetBytes  int
	StrideBytes  int32
}

func NewVertexArrayAttribute(index uint32, compCount int32, compType uint32, norm bool, offsetBytes, bufferBinding uint32) VertexArrayAttribute {
	return VertexArrayAttribute{
		Index:          index,
		ComponentCount: compCount,
		ComponentType:  compType,
		Normalized:     norm,
		OffsetBytes:    offsetBytes,
		BufferBinding:  bufferBinding,
	}
}

type VertexArrayAttribute struct {
	Index          uint32
	ComponentCount int32
	ComponentType  uint32
	Normalized     bool
	OffsetBytes    uint32
	BufferBinding  uint32
}
