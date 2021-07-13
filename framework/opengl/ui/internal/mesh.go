package internal

import (
	"encoding/binary"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/framework/opengl"
)

const (
	positionAttribIndex = 0
	texCoordAttribIndex = 1
	colorAttribIndex    = 2

	positionSize = 2 * 4
	texCoordSize = 2 * 4
	colorSize    = 4 * 1
	stride       = positionSize + texCoordSize + colorSize
)

func NewMesh(vertexCount int) *Mesh {
	data := make([]byte, vertexCount*24)
	return &Mesh{
		vertexData:    data,
		vertexPlotter: buffer.NewPlotter(data, binary.LittleEndian),
		vertexBuffer:  opengl.NewBuffer(),
		vertexArray:   opengl.NewVertexArray(),
	}
}

type Mesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  *opengl.Buffer
	vertexArray   *opengl.VertexArray
}

func (m *Mesh) Allocate() {
	m.vertexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: true,
		Data:    m.vertexData,
	})

	m.vertexArray.Allocate(opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: m.vertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  int32(stride),
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          positionAttribIndex,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    0,
				BufferBinding:  0,
			},
			{
				Index:          texCoordAttribIndex,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    uint32(positionSize),
				BufferBinding:  0,
			},
			{
				Index:          colorAttribIndex,
				ComponentCount: 4,
				ComponentType:  gl.UNSIGNED_BYTE,
				Normalized:     true,
				OffsetBytes:    uint32(positionSize + texCoordSize),
				BufferBinding:  0,
			},
		},
	})
}

func (m *Mesh) Release() {
	m.vertexArray.Release()
	m.vertexBuffer.Release()
}

func (m *Mesh) Update() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.vertexBuffer.Update(opengl.BufferUpdateInfo{
			Data:        m.vertexData[:m.vertexPlotter.Offset()],
			OffsetBytes: 0,
		})
	}
}

func (m *Mesh) Reset() {
	m.vertexOffset = 0
	m.vertexPlotter.Rewind()
}

func (m *Mesh) Offset() int {
	return m.vertexOffset
}

func (m *Mesh) Append(vertex Vertex) {
	m.vertexPlotter.PlotFloat32(vertex.position.X)
	m.vertexPlotter.PlotFloat32(vertex.position.Y)
	m.vertexPlotter.PlotFloat32(vertex.texCoord.X)
	m.vertexPlotter.PlotFloat32(vertex.texCoord.Y)
	m.vertexPlotter.PlotByte(vertex.color.R)
	m.vertexPlotter.PlotByte(vertex.color.G)
	m.vertexPlotter.PlotByte(vertex.color.B)
	m.vertexPlotter.PlotByte(vertex.color.A)
	m.vertexOffset++
}
