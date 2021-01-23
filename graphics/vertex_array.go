package graphics

import (
	"encoding/binary"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/opengl"
)

const (
	CoordAttributeIndex    = 0
	NormalAttributeIndex   = 1
	TangentAttributeIndex  = 2
	TexCoordAttributeIndex = 3
	ColorAttributeIndex    = 4
)

func NewVertexArrayData(vertices, indices int, layout VertexArrayLayout) VertexArrayData {
	verticesDataSize := 0
	if layout.HasCoord {
		verticesDataSize += 3 * 4
	}
	if layout.HasNormal {
		verticesDataSize += 3 * 4
	}
	if layout.HasTangent {
		verticesDataSize += 3 * 4
	}
	if layout.HasTexCoord {
		verticesDataSize += 2 * 4
	}
	if layout.HasColor {
		verticesDataSize += 4 * 4
	}
	verticesDataSize *= vertices
	indicesDataSize := indices * 2

	return VertexArrayData{
		VertexData: make([]byte, verticesDataSize),
		IndexData:  make([]byte, indicesDataSize),
		Layout:     layout,
	}
}

type VertexArrayData struct {
	VertexData []byte
	IndexData  []byte
	Layout     VertexArrayLayout
}

func NewVertexWriter(vad VertexArrayData) *VertexWriter {
	return &VertexWriter{
		plotter: buffer.NewPlotter(vad.VertexData, binary.LittleEndian),
		layout:  vad.Layout,
	}
}

type VertexWriter struct {
	plotter *buffer.Plotter
	layout  VertexArrayLayout
	offset  int
}

func (w *VertexWriter) SetCoord(x, y, z float32) *VertexWriter {
	w.plotter.Seek(w.layout.CoordOffset + w.offset*int(w.layout.CoordStride))
	w.plotter.PlotFloat32(x)
	w.plotter.PlotFloat32(y)
	w.plotter.PlotFloat32(z)
	return w
}

func (w *VertexWriter) Next() *VertexWriter {
	w.offset++
	return w
}

func NewIndexWriter(vad VertexArrayData) *IndexWriter {
	return &IndexWriter{
		plotter: buffer.NewPlotter(vad.IndexData, binary.LittleEndian),
	}
}

type IndexWriter struct {
	plotter *buffer.Plotter
	offset  int
}

func (w *IndexWriter) SetIndex(index uint16) *IndexWriter {
	w.plotter.Seek(w.offset * 2)
	w.plotter.PlotUint16(index)
	return w
}

func (w *IndexWriter) Next() *IndexWriter {
	w.offset++
	return w
}

type VertexArrayLayout struct {
	HasCoord       bool
	CoordOffset    int
	CoordStride    int32
	HasNormal      bool
	NormalOffset   int
	NormalStride   int32
	HasTangent     bool
	TangentOffset  int
	TangentStride  int32
	HasTexCoord    bool
	TexCoordOffset int
	TexCoordStride int32
	HasColor       bool
	ColorOffset    int
	ColorStride    int32
}

func NewVertexArray() *VertexArray {
	return &VertexArray{
		VertexBuffer: new(opengl.Buffer),
		IndexBuffer:  new(opengl.Buffer),
	}
}

type VertexArray struct {
	ID           uint32
	VertexBuffer *opengl.Buffer
	IndexBuffer  *opengl.Buffer
}

func (a *VertexArray) Allocate(data VertexArrayData) error {
	a.VertexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: false,
		Data:    data.VertexData,
	})
	a.IndexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: false,
		Data:    data.IndexData,
	})

	gl.CreateVertexArrays(1, &a.ID)
	if data.Layout.HasCoord {
		gl.EnableVertexArrayAttrib(a.ID, CoordAttributeIndex)
		gl.VertexArrayVertexBuffer(a.ID, CoordAttributeIndex, a.VertexBuffer.ID(), 0, data.Layout.CoordStride)
		gl.VertexArrayAttribFormat(a.ID, CoordAttributeIndex, 3, gl.FLOAT, false, uint32(data.Layout.CoordOffset))
		gl.VertexArrayAttribBinding(a.ID, CoordAttributeIndex, CoordAttributeIndex)
	}
	if data.Layout.HasNormal {
		gl.EnableVertexArrayAttrib(a.ID, NormalAttributeIndex)
		gl.VertexArrayVertexBuffer(a.ID, NormalAttributeIndex, a.VertexBuffer.ID(), 0, data.Layout.NormalStride)
		gl.VertexArrayAttribFormat(a.ID, NormalAttributeIndex, 3, gl.FLOAT, false, uint32(data.Layout.NormalOffset))
		gl.VertexArrayAttribBinding(a.ID, NormalAttributeIndex, NormalAttributeIndex)
	}
	if data.Layout.HasTangent {
		gl.EnableVertexArrayAttrib(a.ID, TangentAttributeIndex)
		gl.VertexArrayVertexBuffer(a.ID, TangentAttributeIndex, a.VertexBuffer.ID(), 0, data.Layout.TangentStride)
		gl.VertexArrayAttribFormat(a.ID, TangentAttributeIndex, 3, gl.FLOAT, false, uint32(data.Layout.TangentOffset))
		gl.VertexArrayAttribBinding(a.ID, TangentAttributeIndex, TangentAttributeIndex)
	}
	if data.Layout.HasTexCoord {
		gl.EnableVertexArrayAttrib(a.ID, TexCoordAttributeIndex)
		gl.VertexArrayVertexBuffer(a.ID, TexCoordAttributeIndex, a.VertexBuffer.ID(), 0, data.Layout.TexCoordStride)
		gl.VertexArrayAttribFormat(a.ID, TexCoordAttributeIndex, 2, gl.FLOAT, false, uint32(data.Layout.TexCoordOffset))
		gl.VertexArrayAttribBinding(a.ID, TexCoordAttributeIndex, TexCoordAttributeIndex)
	}
	if data.Layout.HasColor {
		gl.EnableVertexArrayAttrib(a.ID, ColorAttributeIndex)
		gl.VertexArrayVertexBuffer(a.ID, ColorAttributeIndex, a.VertexBuffer.ID(), 0, data.Layout.ColorStride)
		gl.VertexArrayAttribFormat(a.ID, ColorAttributeIndex, 4, gl.FLOAT, false, uint32(data.Layout.ColorOffset))
		gl.VertexArrayAttribBinding(a.ID, ColorAttributeIndex, ColorAttributeIndex)
	}
	gl.VertexArrayElementBuffer(a.ID, a.IndexBuffer.ID())
	return nil
}

func (a *VertexArray) Release() error {
	a.IndexBuffer.Release()
	a.VertexBuffer.Release()
	gl.DeleteVertexArrays(1, &a.ID)
	a.ID = 0
	return nil
}
