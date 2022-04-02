package internal

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/render"
)

const (
	shapePositionAttribIndex = 0

	shapeMeshVertexSize = 2 * 4
)

func newShapeMesh(vertexCount int) *ShapeMesh {
	data := make([]byte, vertexCount*shapeMeshVertexSize)
	return &ShapeMesh{
		vertexData:    data,
		vertexPlotter: buffer.NewPlotter(data, binary.LittleEndian),
	}
}

type ShapeMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  render.Buffer
	vertexArray   render.VertexArray
}

func (m *ShapeMesh) Allocate(api render.API) {
	m.vertexBuffer = api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: true,
		Data:    m.vertexData,
		// TODO: use Size instead of passing Data
		// Size:    len(m.vertexData),
	})

	m.vertexArray = api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: m.vertexBuffer,
				Stride:       shapeMeshVertexSize,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: shapePositionAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
		},
	})
}

func (m *ShapeMesh) Release() {
	m.vertexArray.Release()
	m.vertexBuffer.Release()
}

func (m *ShapeMesh) Update() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.vertexBuffer.Update(render.BufferUpdateInfo{
			Data:   m.vertexData[:length],
			Offset: 0,
		})
	}
}

func (m *ShapeMesh) Reset() {
	m.vertexOffset = 0
	m.vertexPlotter.Rewind()
}

func (m *ShapeMesh) Offset() int {
	return m.vertexOffset
}

func (m *ShapeMesh) Append(vertex ShapeVertex) {
	m.vertexPlotter.PlotFloat32(vertex.position.X)
	m.vertexPlotter.PlotFloat32(vertex.position.Y)
	m.vertexOffset++
}

type ShapeVertex struct {
	position sprec.Vec2
}
