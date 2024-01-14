package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

type uniformType interface {
	Plot(plotter *blob.Plotter, padding int)
	Std140Size() int
}

func newUniformSequence[T uniformType](api render.API, count int) *uniformSequence[T] {
	var zeroT T
	std140Size := zeroT.Std140Size()
	alignmentSize := render.DetermineUniformBlockSize(api, std140Size)
	data := make([]byte, count*alignmentSize)

	return &uniformSequence[T]{
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

type uniformSequence[T uniformType] struct {
	std140Size    int
	alignmentSize int
	offset        int

	plotter *blob.Plotter
	buffer  render.Buffer
}

func (s *uniformSequence[T]) Buffer() render.Buffer {
	return s.buffer
}

func (s *uniformSequence[T]) BlockOffset() int {
	return s.offset
}

func (s *uniformSequence[T]) BlockSize() int {
	return s.alignmentSize
}

func (s *uniformSequence[T]) Reset() {
	s.offset = 0
	s.plotter.Rewind()
}

func (s *uniformSequence[T]) Append(uniform T) {
	s.offset = s.plotter.Offset()
	uniform.Plot(s.plotter, s.alignmentSize-s.std140Size)
}

func (s *uniformSequence[T]) Upload(api render.API) {
	if offset := s.plotter.Offset(); offset > 0 {
		s.buffer.Update(render.BufferUpdateInfo{
			Data: s.plotter.Data()[:offset],
		})
	}
}

func (s *uniformSequence[T]) Release() {
	s.buffer.Release()
}

type cameraUniform struct {
	Projection sprec.Mat4
}

func (u cameraUniform) Plot(plotter *blob.Plotter, padding int) {
	plotter.PlotSPMat4(u.Projection)
	plotter.Skip(padding)
}

func (u cameraUniform) Std140Size() int {
	return 64
}

type modelUniform struct {
	Transform     sprec.Mat4
	ClipTransform sprec.Mat4
}

func (u modelUniform) Plot(plotter *blob.Plotter, padding int) {
	plotter.PlotSPMat4(u.Transform)
	plotter.PlotSPMat4(u.ClipTransform)
	plotter.Skip(padding)
}

func (u modelUniform) Std140Size() int {
	return 64 + 64
}

type materialUniform struct {
	TextureTransform sprec.Mat4
	Color            sprec.Vec4
}

func (u materialUniform) Plot(plotter *blob.Plotter, padding int) {
	plotter.PlotSPMat4(u.TextureTransform)
	plotter.PlotSPVec4(u.Color)
	plotter.Skip(padding)
}

func (u materialUniform) Std140Size() int {
	return 64 + 16
}
