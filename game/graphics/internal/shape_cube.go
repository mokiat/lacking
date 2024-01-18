package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

func CreateCubeShape(api render.API) *Shape {
	const (
		vertexSize  = 3 * render.SizeF32
		vertexCount = 8

		indexSize  = 1 * render.SizeU16
		indexCount = 36
	)

	vertexData := make([]byte, vertexCount*vertexSize)
	vertexPlotter := blob.NewPlotter(vertexData)
	vertexPlotter.PlotSPVec3(sprec.NewVec3(-1.0, 1.0, 1.0))
	vertexPlotter.PlotSPVec3(sprec.NewVec3(-1.0, -1.0, 1.0))
	vertexPlotter.PlotSPVec3(sprec.NewVec3(1.0, -1.0, 1.0))
	vertexPlotter.PlotSPVec3(sprec.NewVec3(1.0, 1.0, 1.0))

	vertexPlotter.PlotSPVec3(sprec.NewVec3(-1.0, 1.0, -1.0))
	vertexPlotter.PlotSPVec3(sprec.NewVec3(-1.0, -1.0, -1.0))
	vertexPlotter.PlotSPVec3(sprec.NewVec3(1.0, -1.0, -1.0))
	vertexPlotter.PlotSPVec3(sprec.NewVec3(1.0, 1.0, -1.0))

	indexData := make([]byte, indexCount*indexSize)
	indexPlotter := blob.NewPlotter(indexData)
	indexPlotter.PlotUint16(3)
	indexPlotter.PlotUint16(1)
	indexPlotter.PlotUint16(2)

	indexPlotter.PlotUint16(3)
	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(1)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(5)
	indexPlotter.PlotUint16(1)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(4)
	indexPlotter.PlotUint16(5)

	indexPlotter.PlotUint16(7)
	indexPlotter.PlotUint16(2)
	indexPlotter.PlotUint16(6)

	indexPlotter.PlotUint16(7)
	indexPlotter.PlotUint16(3)
	indexPlotter.PlotUint16(2)

	indexPlotter.PlotUint16(4)
	indexPlotter.PlotUint16(6)
	indexPlotter.PlotUint16(5)

	indexPlotter.PlotUint16(4)
	indexPlotter.PlotUint16(7)
	indexPlotter.PlotUint16(6)

	indexPlotter.PlotUint16(5)
	indexPlotter.PlotUint16(2)
	indexPlotter.PlotUint16(1)

	indexPlotter.PlotUint16(5)
	indexPlotter.PlotUint16(6)
	indexPlotter.PlotUint16(2)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(7)
	indexPlotter.PlotUint16(4)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(3)
	indexPlotter.PlotUint16(7)

	vertexBuffer := api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    vertexData,
	})

	indexBuffer := api.CreateIndexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    indexData,
	})

	vertexArray := api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBinding{
			{
				VertexBuffer: vertexBuffer,
				Stride:       vertexSize,
			},
		},
		Attributes: []render.VertexArrayAttribute{
			{
				Binding:  0,
				Location: CoordAttributeIndex,
				Format:   render.VertexAttributeFormatRGB32F,
				Offset:   0,
			},
		},
		IndexBuffer: indexBuffer,
		IndexFormat: render.IndexFormatUnsignedShort,
	})

	return &Shape{
		vertexBuffer: vertexBuffer,
		indexBuffer:  indexBuffer,

		vertexArray: vertexArray,

		topology:   render.TopologyTriangleList,
		indexCount: indexCount,
	}
}
