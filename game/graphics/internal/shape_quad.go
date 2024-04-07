package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

func CreateQuadShape(api render.API) *Shape {
	const (
		vertexSize  = 2 * render.SizeF32
		vertexCount = 4

		indexSize  = 1 * render.SizeU16
		indexCount = 6
	)

	vertexData := make([]byte, vertexCount*vertexSize)
	vertexPlotter := blob.NewPlotter(vertexData)
	vertexPlotter.PlotSPVec2(sprec.NewVec2(-1.0, 1.0))
	vertexPlotter.PlotSPVec2(sprec.NewVec2(-1.0, -1.0))
	vertexPlotter.PlotSPVec2(sprec.NewVec2(1.0, -1.0))
	vertexPlotter.PlotSPVec2(sprec.NewVec2(1.0, 1.0))

	indexData := make([]byte, indexCount*indexSize)
	indexPlotter := blob.NewPlotter(indexData)
	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(1)
	indexPlotter.PlotUint16(2)

	indexPlotter.PlotUint16(0)
	indexPlotter.PlotUint16(2)
	indexPlotter.PlotUint16(3)

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
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
		},
		IndexBuffer: indexBuffer,
		IndexFormat: render.IndexFormatUnsignedU16,
	})

	return &Shape{
		vertexBuffer: vertexBuffer,
		indexBuffer:  indexBuffer,

		vertexArray: vertexArray,

		topology:   render.TopologyTriangleList,
		indexCount: indexCount,
	}
}
