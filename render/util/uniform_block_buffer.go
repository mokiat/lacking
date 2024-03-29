package util

import (
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

// UniformType represents a type that can be written to a uniform buffer.
type UniformType interface {
	Std140Plot(plotter *blob.Plotter)
	Std140Size() int
}

// UniformPlacement contains information on the positioning of a uniform
// within a uniform buffer.
type UniformPlacement struct {
	Buffer  render.Buffer
	Plotter *blob.Plotter
	Offset  int
	Size    int
}

// WriteUniform writes the specified uniform to the specified uniform block
// buffer. The returned placement can be used to retrieve the uniform's
// position within the buffer.
func WriteUniform[T UniformType](blockBuffer *UniformBlockBuffer, uniform T) UniformPlacement {
	size := uniform.Std140Size()
	placement := blockBuffer.Placement(size)
	uniform.Std140Plot(placement.Plotter)
	placement.Plotter = nil // prevent writing
	return placement
}

// NewUniformBlockBuffer creates a new uniform block buffer that can be
// used to store multiple uniform block types.
//
// NOTE: This does not do automatic resize so the proper capacity needs
// to be figured out beforehand.
func NewUniformBlockBuffer(api render.API, capacity int) *UniformBlockBuffer {
	data := make([]byte, capacity)
	return &UniformBlockBuffer{
		api:     api,
		plotter: blob.NewPlotter(data),
		buffer: api.CreateUniformBuffer(render.BufferInfo{
			Dynamic: true,
			Size:    len(data),
		}),
		blockAlignment: api.Limits().UniformBufferOffsetAlignment(),
	}
}

// UniformBlockBuffer represents a shared uniform buffer that can be used
// for storing multiple uniform block types.
type UniformBlockBuffer struct {
	// IDEA: A future version could have a sequence of uniform buffers and
	// add to them as needed.

	api            render.API
	plotter        *blob.Plotter
	buffer         render.Buffer
	blockAlignment int
}

// Reset resets the uniform block buffer so that it can be reused.
func (b *UniformBlockBuffer) Reset() {
	b.plotter.Rewind()
}

// Placement returns a UniformPlacement that can be used to write a uniform
// of the specified size.
func (b *UniformBlockBuffer) Placement(uniformSize int) UniformPlacement {
	b.skipToAlignment()
	return UniformPlacement{
		Buffer:  b.buffer,
		Plotter: b.plotter,
		Offset:  b.plotter.Offset(),
		Size:    uniformSize,
	}
}

// Upload uploads the data that has been written to the uniform block buffer
// to the GPU.
func (b *UniformBlockBuffer) Upload() {
	b.skipToAlignment()
	if offset := b.plotter.Offset(); offset > 0 {
		b.api.Queue().WriteBuffer(b.buffer, 0, b.plotter.Data()[:offset])
	}
}

// Release releases the resources associated with the uniform block buffer.
func (b *UniformBlockBuffer) Release() {
	b.buffer.Release()
}

func (b *UniformBlockBuffer) skipToAlignment() {
	offset := b.plotter.Offset()
	if overshoot := offset % b.blockAlignment; overshoot > 0 {
		b.plotter.Skip(b.blockAlignment - overshoot)
	}
}
