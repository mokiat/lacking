package internal

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/render"
)

const (
	textPositionAttribIndex = 0
	textTexCoordAttribIndex = 1

	textMeshVertexSize = 2*4 + 2*4
)

func newTextMesh(vertexCount int) *TextMesh {
	data := make([]byte, vertexCount*textMeshVertexSize)
	return &TextMesh{
		vertexData:    data,
		vertexPlotter: buffer.NewPlotter(data, binary.LittleEndian),
	}
}

type TextMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  render.Buffer
	vertexArray   render.VertexArray
}

func (m *TextMesh) Allocate(api render.API) {
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
				Stride:       textMeshVertexSize,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: textPositionAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
			{
				Binding:  0,
				Location: textTexCoordAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   2 * 4,
			},
		},
	})
}

func (m *TextMesh) Release() {
	m.vertexArray.Release()
	m.vertexBuffer.Release()
}

func (m *TextMesh) Update() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.vertexBuffer.Update(render.BufferUpdateInfo{
			Data:   m.vertexData[:length],
			Offset: 0,
		})
	}
}

func (m *TextMesh) Reset() {
	m.vertexOffset = 0
	m.vertexPlotter.Rewind()
}

func (m *TextMesh) Offset() int {
	return m.vertexOffset
}

func (m *TextMesh) Append(vertex TextVertex) {
	m.vertexPlotter.PlotFloat32(vertex.position.X)
	m.vertexPlotter.PlotFloat32(vertex.position.Y)
	m.vertexPlotter.PlotFloat32(vertex.texCoord.X)
	m.vertexPlotter.PlotFloat32(vertex.texCoord.Y)
	m.vertexOffset++
}

type TextVertex struct {
	position sprec.Vec2
	texCoord sprec.Vec2
}
