package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

func CreateSphereShape(api render.API) *Shape {
	const (
		slices               = 9
		horizontalSliceCount = 2 + 2*slices
		verticalSliceCount   = 2 + slices

		vertexSize  = 3 * render.SizeF32
		vertexCount = 2 * (1 + slices*(1+slices))

		indexSize  = 1 * render.SizeU16
		indexCount = (3 + 6*(slices-1) + 3) * horizontalSliceCount
	)

	vertexData := make([]byte, vertexCount*vertexSize)
	vertexPlotter := blob.NewPlotter(vertexData)
	vertexPlotter.PlotSPVec3(sprec.NewVec3(0.0, 1.0, 0.0))  // top
	vertexPlotter.PlotSPVec3(sprec.NewVec3(0.0, -1.0, 0.0)) // bottom
	for hs := 0; hs < horizontalSliceCount; hs++ {
		hAngle := sprec.Radians(2 * sprec.Pi * (float32(-hs) / float32(horizontalSliceCount)))
		hCos := sprec.Cos(hAngle)
		hSin := sprec.Sin(hAngle)
		for vs := 1; vs <= slices; vs++ {
			vAngle := sprec.Radians(sprec.Pi/2.0 - sprec.Pi*(float32(vs)/float32(verticalSliceCount-1)))
			vCos := sprec.Cos(vAngle)
			vSin := sprec.Sin(vAngle)
			vertexPlotter.PlotSPVec3(sprec.NewVec3(hCos*vCos, vSin, hSin*vCos))
		}
	}

	indexData := make([]byte, indexCount*indexSize)
	indexPlotter := blob.NewPlotter(indexData)
	for x := 0; x < horizontalSliceCount; x++ {
		left := x % horizontalSliceCount
		right := (x + 1) % horizontalSliceCount
		leftOffset := left * slices
		rightOffset := right * slices

		upperLeft := uint16(2 + leftOffset)
		upperRight := uint16(2 + rightOffset)

		indexPlotter.PlotUint16(0) // top
		indexPlotter.PlotUint16(upperLeft)
		indexPlotter.PlotUint16(upperRight)

		for y := 0; y < slices-1; y++ {
			lowerLeft := upperLeft + 1
			lowerRight := upperRight + 1

			indexPlotter.PlotUint16(upperLeft)
			indexPlotter.PlotUint16(lowerLeft)
			indexPlotter.PlotUint16(lowerRight)

			indexPlotter.PlotUint16(upperLeft)
			indexPlotter.PlotUint16(lowerRight)
			indexPlotter.PlotUint16(upperRight)

			upperLeft++
			upperRight++
		}

		indexPlotter.PlotUint16(upperLeft)
		indexPlotter.PlotUint16(1) // bottom
		indexPlotter.PlotUint16(upperRight)
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
