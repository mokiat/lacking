package internal

import (
	"encoding/binary"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/framework/opengl"
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
		vertexBuffer:  opengl.NewBuffer(),
		vertexArray:   opengl.NewVertexArray(),
	}
}

type TextMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  *opengl.Buffer
	vertexArray   *opengl.VertexArray
}

func (m *TextMesh) Allocate() {
	m.vertexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: true,
		Data:    m.vertexData,
	})

	m.vertexArray.Allocate(opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: m.vertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  int32(textMeshVertexSize),
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          textPositionAttribIndex,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    0,
				BufferBinding:  0,
			},
			{
				Index:          textTexCoordAttribIndex,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    2 * 4,
				BufferBinding:  0,
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
		m.vertexBuffer.Update(opengl.BufferUpdateInfo{
			Data:        m.vertexData[:m.vertexPlotter.Offset()],
			OffsetBytes: 0,
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
