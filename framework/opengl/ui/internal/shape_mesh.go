package internal

import (
	"encoding/binary"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/framework/opengl"
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
		vertexBuffer:  opengl.NewBuffer(),
		vertexArray:   opengl.NewVertexArray(),
	}
}

type ShapeMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  *opengl.Buffer
	vertexArray   *opengl.VertexArray
}

func (m *ShapeMesh) Allocate() {
	m.vertexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: true,
		Data:    m.vertexData,
	})

	m.vertexArray.Allocate(opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: m.vertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  int32(shapeMeshVertexSize),
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          shapePositionAttribIndex,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    0,
				BufferBinding:  0,
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
		m.vertexBuffer.Update(opengl.BufferUpdateInfo{
			Data:        m.vertexData[:m.vertexPlotter.Offset()],
			OffsetBytes: 0,
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
