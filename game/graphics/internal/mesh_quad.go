package internal

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/render"
)

type quadMeshVertex struct {
	Position sprec.Vec2
}

func (v quadMeshVertex) Serialize(plotter *buffer.Plotter) {
	plotter.PlotFloat32(v.Position.X)
	plotter.PlotFloat32(v.Position.Y)
}

func NewQuadMesh() *QuadMesh {
	return &QuadMesh{
		Topology:         render.TopologyTriangles,
		IndexCount:       6,
		IndexOffsetBytes: 0,
	}
}

type QuadMesh struct {
	Topology         render.Topology
	IndexCount       int
	IndexOffsetBytes int

	VertexBuffer render.Buffer
	IndexBuffer  render.Buffer
	VertexArray  render.VertexArray
}

func (m *QuadMesh) Allocate(api render.API) {
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
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
		},
		IndexBuffer: m.IndexBuffer,
		IndexFormat: render.IndexFormatUnsignedShort,
	})
}

func (m *QuadMesh) Release() {
	m.VertexArray.Release()
	m.IndexBuffer.Release()
	m.VertexBuffer.Release()
}
