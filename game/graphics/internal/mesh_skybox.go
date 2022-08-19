package internal

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/render"
)

type skyboxMeshVertex struct {
	Position sprec.Vec3
}

func (v skyboxMeshVertex) Serialize(plotter *buffer.Plotter) {
	plotter.PlotFloat32(v.Position.X)
	plotter.PlotFloat32(v.Position.Y)
	plotter.PlotFloat32(v.Position.Z)
}

func NewSkyboxMesh() *SkyboxMesh {
	return &SkyboxMesh{
		Topology:         render.TopologyTriangles,
		IndexCount:       36,
		IndexOffsetBytes: 0,
	}
}

type SkyboxMesh struct {
	Topology         render.Topology
	IndexCount       int
	IndexOffsetBytes int

	VertexBuffer render.Buffer
	IndexBuffer  render.Buffer
	VertexArray  render.VertexArray
}

func (m *SkyboxMesh) Allocate(api render.API) {
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

	m.VertexBuffer = api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    vertexPlotter.Data(),
	})

	m.IndexBuffer = api.CreateIndexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    indexPlotter.Data(),
	})

	m.VertexArray = api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: m.VertexBuffer,
				Stride:       vertexSize,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: CoordAttributeIndex,
				Format:   render.VertexAttributeFormatRGB32F,
				Offset:   0,
			},
		},
		IndexBuffer: m.IndexBuffer,
		IndexFormat: render.IndexFormatUnsignedShort,
	})
}

func (m *SkyboxMesh) Release() {
	m.VertexArray.Release()
	m.IndexBuffer.Release()
	m.VertexBuffer.Release()
}
