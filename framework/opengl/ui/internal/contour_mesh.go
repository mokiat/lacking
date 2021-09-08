package internal

import (
	"encoding/binary"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/framework/opengl"
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
		vertexBuffer:  opengl.NewBuffer(),
		vertexArray:   opengl.NewVertexArray(),
	}
}

type ContourMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  *opengl.Buffer
	vertexArray   *opengl.VertexArray
}

func (m *ContourMesh) Allocate() {
	m.vertexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: true,
		Data:    m.vertexData,
	})

	m.vertexArray.Allocate(opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: m.vertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  int32(contourMeshVertexSize),
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          contourPositionAttribIndex,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    0,
				BufferBinding:  0,
			},
			{
				Index:          contourColorAttribIndex,
				ComponentCount: 4,
				ComponentType:  gl.UNSIGNED_BYTE,
				Normalized:     false,
				OffsetBytes:    2 * 4,
				BufferBinding:  0,
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
		m.vertexBuffer.Update(opengl.BufferUpdateInfo{
			Data:        m.vertexData[:m.vertexPlotter.Offset()],
			OffsetBytes: 0,
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
