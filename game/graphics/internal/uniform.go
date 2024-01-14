package internal

import (
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

type UniformType interface {
	Plot(plotter *blob.Plotter, padding int)
	Std140Size() int
}

func NewUniformSequence[T UniformType](api render.API, count int) *UniformSequence[T] {
	var zeroT T
	std140Size := zeroT.Std140Size()
	alignmentSize := render.DetermineUniformBlockSize(api, std140Size)
	data := make([]byte, count*alignmentSize)

	return &UniformSequence[T]{
		std140Size:    std140Size,
		alignmentSize: alignmentSize,
		offset:        0,

		plotter: blob.NewPlotter(data),
		buffer: api.CreateUniformBuffer(render.BufferInfo{
			Dynamic: true,
			Size:    len(data),
		}),
	}
}

type UniformSequence[T UniformType] struct {
	std140Size    int
	alignmentSize int
	offset        int

	plotter *blob.Plotter
	buffer  render.Buffer
}

func (s *UniformSequence[T]) Buffer() render.Buffer {
	return s.buffer
}

func (s *UniformSequence[T]) BlockOffset() int {
	return s.offset
}

func (s *UniformSequence[T]) BlockSize() int {
	return s.alignmentSize
}

func (s *UniformSequence[T]) Reset() {
	s.offset = 0
	s.plotter.Rewind()
}

func (s *UniformSequence[T]) Append(uniform T) {
	s.offset = s.plotter.Offset()
	uniform.Plot(s.plotter, s.alignmentSize-s.std140Size)
}

func (s *UniformSequence[T]) Upload(api render.API) {
	s.buffer.Update(render.BufferUpdateInfo{
		Data: s.plotter.Data()[:s.plotter.Offset()],
	})
}

func (s *UniformSequence[T]) Release() {
	s.buffer.Release()
}
