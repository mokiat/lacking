package graphics

import (
	"encoding/binary"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/framework/opengl"
)

type skyboxMeshVertex struct {
	Position sprec.Vec3
}

func (v skyboxMeshVertex) Serialize(plotter *buffer.Plotter) {
	plotter.PlotFloat32(v.Position.X)
	plotter.PlotFloat32(v.Position.Y)
	plotter.PlotFloat32(v.Position.Z)
}

func newSkyboxMesh() *SkyboxMesh {
	return &SkyboxMesh{
		VertexBuffer:     opengl.NewBuffer(),
		IndexBuffer:      opengl.NewBuffer(),
		VertexArray:      opengl.NewVertexArray(),
		Primitive:        gl.TRIANGLES,
		IndexCount:       36,
		IndexOffsetBytes: 0,
	}
}

type SkyboxMesh struct {
	VertexBuffer     *opengl.Buffer
	IndexBuffer      *opengl.Buffer
	VertexArray      *opengl.VertexArray
	Primitive        uint32
	IndexCount       int32
	IndexOffsetBytes int
}

func (m *SkyboxMesh) Allocate() {
	const vertexSize = 3 * 4
	vertexPlotter := buffer.NewPlotter(
		make([]byte, vertexSize*8),
		binary.LittleEndian,
	)

	skyboxMeshVertex{
		Position: sprec.NewVec3(-1.0, 1.0, 1.0),
	}.Serialize(vertexPlotter)
	skyboxMeshVertex{
		Position: sprec.NewVec3(-1.0, -1.0, 1.0),
	}.Serialize(vertexPlotter)
	skyboxMeshVertex{
		Position: sprec.NewVec3(1.0, -1.0, 1.0),
	}.Serialize(vertexPlotter)
	skyboxMeshVertex{
		Position: sprec.NewVec3(1.0, 1.0, 1.0),
	}.Serialize(vertexPlotter)

	skyboxMeshVertex{
		Position: sprec.NewVec3(-1.0, 1.0, -1.0),
	}.Serialize(vertexPlotter)
	skyboxMeshVertex{
		Position: sprec.NewVec3(-1.0, -1.0, -1.0),
	}.Serialize(vertexPlotter)
	skyboxMeshVertex{
		Position: sprec.NewVec3(1.0, -1.0, -1.0),
	}.Serialize(vertexPlotter)
	skyboxMeshVertex{
		Position: sprec.NewVec3(1.0, 1.0, -1.0),
	}.Serialize(vertexPlotter)

	const indexSize = 1 * 2
	indexPlotter := buffer.NewPlotter(
		make([]byte, indexSize*36),
		binary.LittleEndian,
	)

	indexPlotter.PlotUint16(3)
	indexPlotter.PlotUint16(2)
	indexPlotter.PlotUint16(1)

	indexPlotter.PlotUint16(3)
	indexPlotter.PlotUint16(1)
	indexPlotter.PlotUint16(0)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(1)
	indexPlotter.PlotUint16(5)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(5)
	indexPlotter.PlotUint16(4)

	indexPlotter.PlotUint16(7)
	indexPlotter.PlotUint16(6)
	indexPlotter.PlotUint16(2)

	indexPlotter.PlotUint16(7)
	indexPlotter.PlotUint16(2)
	indexPlotter.PlotUint16(3)

	indexPlotter.PlotUint16(4)
	indexPlotter.PlotUint16(5)
	indexPlotter.PlotUint16(6)

	indexPlotter.PlotUint16(4)
	indexPlotter.PlotUint16(6)
	indexPlotter.PlotUint16(7)

	indexPlotter.PlotUint16(5)
	indexPlotter.PlotUint16(1)
	indexPlotter.PlotUint16(2)

	indexPlotter.PlotUint16(5)
	indexPlotter.PlotUint16(2)
	indexPlotter.PlotUint16(6)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(4)
	indexPlotter.PlotUint16(7)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(7)
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
				StrideBytes:  3 * 4,
			},
		},
		Attributes: []opengl.VertexArrayAttribute{
			{
				Index:          0,
				ComponentCount: 3,
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

func (m *SkyboxMesh) Release() {
	m.VertexArray.Release()
	m.IndexBuffer.Release()
	m.VertexBuffer.Release()
}
