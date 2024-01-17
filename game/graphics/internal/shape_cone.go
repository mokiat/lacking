package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

func CreateConeShape(api render.API) *Shape {
	const (
		slices = 12

		vertexSize  = 3 * render.SizeF32
		vertexCount = 1 + slices

		indexSize  = 1 * render.SizeU16
		indexCount = (1 + slices) + 1 + (1 + slices)
	)

	vertexData := make([]byte, vertexCount*vertexSize)
	vertexPlotter := blob.NewPlotter(vertexData)
	vertexPlotter.PlotSPVec3(sprec.NewVec3(0.0, 1.0, 0.0)) // top
	for s := 0; s < slices; s++ {
		angle := sprec.Radians(2 * sprec.Pi * (float32(s) / float32(slices)))
		cs := sprec.Cos(angle)
		sn := sprec.Sin(angle)
		vertexPlotter.PlotSPVec3(sprec.NewVec3(cs, 0.0, -sn))
	}

	indexData := make([]byte, indexCount*indexSize)
	indexPlotter := blob.NewPlotter(indexData)
	// sides
	indexPlotter.PlotUint16(0)
	for s := 1; s <= slices; s++ {
		indexPlotter.PlotUint16(uint16(s))
	}
	indexPlotter.PlotUint16(1)
	// bottom
	indexPlotter.PlotUint16(0xFFFF)
	for s := slices; s >= 1; s-- {
		indexPlotter.PlotUint16(uint16(s))
	}

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

		topology:   render.TopologyTriangleFan,
		indexCount: indexCount,
	}
}
