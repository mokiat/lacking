package preset

import "github.com/mokiat/lacking/graphics"

func NewQuadVertexArrayData(left, right, top, bottom float32) graphics.VertexArrayData {
	result := graphics.NewVertexArrayData(4, 6, graphics.VertexArrayLayout{
		HasCoord:    true,
		CoordOffset: 0,
		CoordStride: 3 * 4,
	})

	vertexWriter := graphics.NewVertexWriter(result)
	vertexWriter.SetCoord(left, top, 0.0).Next()
	vertexWriter.SetCoord(left, bottom, 0.0).Next()
	vertexWriter.SetCoord(right, bottom, 0.0).Next()
	vertexWriter.SetCoord(right, top, 0.0).Next()

	indexWriter := graphics.NewIndexWriter(result)
	indexWriter.SetIndex(0).Next()
	indexWriter.SetIndex(1).Next()
	indexWriter.SetIndex(2).Next()
	indexWriter.SetIndex(0).Next()
	indexWriter.SetIndex(2).Next()
	indexWriter.SetIndex(3).Next()

	return result
}

func NewCubeVertexArrayData(left, right, top, bottom, near, far float32) graphics.VertexArrayData {
	result := graphics.NewVertexArrayData(8, 36, graphics.VertexArrayLayout{
		HasCoord:    true,
		CoordOffset: 0,
		CoordStride: 3 * 4,
	})

	vertexWriter := graphics.NewVertexWriter(result)
	vertexWriter.SetCoord(-1.0, 1.0, 1.0).Next()
	vertexWriter.SetCoord(-1.0, -1.0, 1.0).Next()
	vertexWriter.SetCoord(1.0, -1.0, 1.0).Next()
	vertexWriter.SetCoord(1.0, 1.0, 1.0).Next()

	vertexWriter.SetCoord(-1.0, 1.0, -1.0).Next()
	vertexWriter.SetCoord(-1.0, -1.0, -1.0).Next()
	vertexWriter.SetCoord(1.0, -1.0, -1.0).Next()
	vertexWriter.SetCoord(1.0, 1.0, -1.0).Next()

	indexWriter := graphics.NewIndexWriter(result)
	indexWriter.SetIndex(3).Next()
	indexWriter.SetIndex(2).Next()
	indexWriter.SetIndex(1).Next()

	indexWriter.SetIndex(3).Next()
	indexWriter.SetIndex(1).Next()
	indexWriter.SetIndex(0).Next()

	indexWriter.SetIndex(0).Next()
	indexWriter.SetIndex(1).Next()
	indexWriter.SetIndex(5).Next()

	indexWriter.SetIndex(0).Next()
	indexWriter.SetIndex(5).Next()
	indexWriter.SetIndex(4).Next()

	indexWriter.SetIndex(7).Next()
	indexWriter.SetIndex(6).Next()
	indexWriter.SetIndex(2).Next()

	indexWriter.SetIndex(7).Next()
	indexWriter.SetIndex(2).Next()
	indexWriter.SetIndex(3).Next()

	indexWriter.SetIndex(4).Next()
	indexWriter.SetIndex(5).Next()
	indexWriter.SetIndex(6).Next()

	indexWriter.SetIndex(4).Next()
	indexWriter.SetIndex(6).Next()
	indexWriter.SetIndex(7).Next()

	indexWriter.SetIndex(5).Next()
	indexWriter.SetIndex(1).Next()
	indexWriter.SetIndex(2).Next()

	indexWriter.SetIndex(5).Next()
	indexWriter.SetIndex(2).Next()
	indexWriter.SetIndex(6).Next()

	indexWriter.SetIndex(0).Next()
	indexWriter.SetIndex(4).Next()
	indexWriter.SetIndex(7).Next()

	indexWriter.SetIndex(0).Next()
	indexWriter.SetIndex(7).Next()
	indexWriter.SetIndex(3).Next()

	return result
}
