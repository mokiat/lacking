package internal

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/render"
)

const (
	contourPositionAttribIndex = 0
	contourColorAttribIndex    = 2

	contourMeshVertexSize = 2*4 + 1*4
)

func newContourMesh(vertexCount int) *ContourMesh {
	data := make([]byte, vertexCount*contourMeshVertexSize)
	return &ContourMesh{
		vertexData:    data,
		vertexPlotter: buffer.NewPlotter(data, binary.LittleEndian),
	}
}

type ContourMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  render.Buffer
	vertexArray   render.VertexArray
}

func (m *ContourMesh) Allocate(api render.API) {
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
				Stride:       contourMeshVertexSize,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: contourPositionAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
			{
				Binding:  0,
				Location: contourColorAttribIndex,
				Format:   render.VertexAttributeFormatRGBA8UN,
				Offset:   2 * 4,
			},
		},
	})
}

func (m *ContourMesh) Release() {
	m.vertexArray.Release()
	m.vertexBuffer.Release()
}

func (m *ContourMesh) Update() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.vertexBuffer.Update(render.BufferUpdateInfo{
			Data:   m.vertexData[:length],
			Offset: 0,
		})
	}
}

func (m *ContourMesh) Reset() {
	m.vertexOffset = 0
	m.vertexPlotter.Rewind()
}

func (m *ContourMesh) Offset() int {
	return m.vertexOffset
}

func (m *ContourMesh) Append(vertex ContourVertex) {
	m.vertexPlotter.PlotFloat32(vertex.position.X)
	m.vertexPlotter.PlotFloat32(vertex.position.Y)
	m.vertexPlotter.PlotByte(byte(vertex.color.X * 255))
	m.vertexPlotter.PlotByte(byte(vertex.color.Y * 255))
	m.vertexPlotter.PlotByte(byte(vertex.color.Z * 255))
	m.vertexPlotter.PlotByte(byte(vertex.color.W * 255))
	m.vertexOffset++
}

type ContourVertex struct {
	position sprec.Vec2
	color    sprec.Vec4
}
