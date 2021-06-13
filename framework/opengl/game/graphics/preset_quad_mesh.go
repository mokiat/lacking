package graphics

import (
	"encoding/binary"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/framework/opengl"
)

type quadMeshVertex struct {
	Position sprec.Vec2
}

func (v quadMeshVertex) Serialize(plotter *buffer.Plotter) {
	plotter.PlotFloat32(v.Position.X)
	plotter.PlotFloat32(v.Position.Y)
}

func newQuadMesh() *QuadMesh {
	return &QuadMesh{
		VertexBuffer:     opengl.NewBuffer(),
		IndexBuffer:      opengl.NewBuffer(),
		VertexArray:      opengl.NewVertexArray(),
		Primitive:        gl.TRIANGLES,
		IndexCount:       6,
		IndexOffsetBytes: 0,
	}
}

type QuadMesh struct {
	VertexBuffer     *opengl.Buffer
	IndexBuffer      *opengl.Buffer
	VertexArray      *opengl.VertexArray
	Primitive        uint32
	IndexCount       int32
	IndexOffsetBytes int
}

func (m *QuadMesh) Allocate() {
	const vertexSize = 2 * 4
	vertexPlotter := buffer.NewPlotter(
		make([]byte, vertexSize*4),
		binary.LittleEndian,
	)

	quadMeshVertex{
		Position: sprec.NewVec2(-1.0, 1.0),
	}.Serialize(vertexPlotter)
	quadMeshVertex{
		Position: sprec.NewVec2(-1.0, -1.0),
	}.Serialize(vertexPlotter)
	quadMeshVertex{
		Position: sprec.NewVec2(1.0, -1.0),
	}.Serialize(vertexPlotter)
	quadMeshVertex{
		Position: sprec.NewVec2(1.0, 1.0),
	}.Serialize(vertexPlotter)

	const indexSize = 1 * 2
	indexPlotter := buffer.NewPlotter(
		make([]byte, indexSize*6),
		binary.LittleEndian,
	)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(1)
	indexPlotter.PlotUint16(2)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(2)
	indexPlotter.PlotUint16(3)

	vertexBufferInfo := opengl.BufferAllocateInfo{
		Dynamic: false,
		Data:    vertexPlotter.Data(),
	}
	m.VertexBuffer.Allocate(vertexBufferInfo)
	indexBufferInfo := opengl.BufferAllocateInfo{
		Dynamic: false,
		Data:    indexPlotter.Data(),
	}
	m.IndexBuffer.Allocate(indexBufferInfo)
	vertexArrayInfo := opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: m.VertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  vertexSize,
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          0,
				ComponentCount: 2,
				ComponentType:  gl.FLOAT,
				Normalized:     false,
				OffsetBytes:    0,
				BufferBinding:  0,
			},
		},
		IndexBuffer: m.IndexBuffer,
	}
	m.VertexArray.Allocate(vertexArrayInfo)
}

func (m *QuadMesh) Release() {
	m.VertexArray.Release()
	m.IndexBuffer.Release()
	m.VertexBuffer.Release()
}
