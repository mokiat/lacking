package graphics

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/framework/opengl"
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
		vertexData: data.Buffer(vad.VertexData),
		layout:     vad.Layout,
	}
}

type VertexWriter struct {
	vertexData data.Buffer
	layout     VertexArrayLayout
	offset     int
}

func (w *VertexWriter) SetCoord(x, y, z float32) *VertexWriter {
	offset := w.layout.CoordOffset + w.offset*int(w.layout.CoordStride)
	w.vertexData.SetFloat32(offset+0, x)
	w.vertexData.SetFloat32(offset+4, y)
	w.vertexData.SetFloat32(offset+8, z)
	return w
}

func (w *VertexWriter) Next() *VertexWriter {
	w.offset++
	return w
}

func NewIndexWriter(vad VertexArrayData) *IndexWriter {
	return &IndexWriter{
		indexData: data.Buffer(vad.IndexData),
	}
}

type IndexWriter struct {
	indexData data.Buffer
	offset    int
}

func (w *IndexWriter) SetIndex(index uint16) *IndexWriter {
	w.indexData.SetUInt16(w.offset*2, index)
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
		VertexArray:  opengl.NewVertexArray(),
		VertexBuffer: opengl.NewBuffer(),
		IndexBuffer:  opengl.NewBuffer(),
	}
}

type VertexArray struct {
	VertexArray  *opengl.VertexArray
	VertexBuffer *opengl.Buffer
	IndexBuffer  *opengl.Buffer
}

func (a *VertexArray) ID() uint32 {
	return a.VertexArray.ID()
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

	vertexArrayInfo := opengl.VertexArrayAllocateInfo{
		IndexBuffer: a.IndexBuffer,
	}
	if data.Layout.HasCoord {
		attribute := opengl.NewVertexArrayAttribute(
			CoordAttributeIndex,
			3, gl.FLOAT, false,
			uint32(data.Layout.CoordOffset),
			uint32(len(vertexArrayInfo.BufferBindings)),
		)
		binding := opengl.NewVertexArrayBufferBinding(
			a.VertexBuffer,
			0, data.Layout.CoordStride,
		)
		vertexArrayInfo.Attributes = append(vertexArrayInfo.Attributes, attribute)
		vertexArrayInfo.BufferBindings = append(vertexArrayInfo.BufferBindings, binding)
	}
	if data.Layout.HasNormal {
		attribute := opengl.NewVertexArrayAttribute(
			NormalAttributeIndex,
			3, gl.FLOAT, false,
			uint32(data.Layout.NormalOffset),
			uint32(len(vertexArrayInfo.BufferBindings)),
		)
		binding := opengl.NewVertexArrayBufferBinding(
			a.VertexBuffer,
			0, data.Layout.NormalStride,
		)
		vertexArrayInfo.Attributes = append(vertexArrayInfo.Attributes, attribute)
		vertexArrayInfo.BufferBindings = append(vertexArrayInfo.BufferBindings, binding)
	}
	if data.Layout.HasTangent {
		attribute := opengl.NewVertexArrayAttribute(
			TangentAttributeIndex,
			3, gl.FLOAT, false,
			uint32(data.Layout.TangentOffset),
			uint32(len(vertexArrayInfo.BufferBindings)),
		)
		binding := opengl.NewVertexArrayBufferBinding(
			a.VertexBuffer,
			0, data.Layout.TangentStride,
		)
		vertexArrayInfo.Attributes = append(vertexArrayInfo.Attributes, attribute)
		vertexArrayInfo.BufferBindings = append(vertexArrayInfo.BufferBindings, binding)
	}
	if data.Layout.HasTexCoord {
		attribute := opengl.NewVertexArrayAttribute(
			TexCoordAttributeIndex,
			2, gl.FLOAT, false,
			uint32(data.Layout.TexCoordOffset),
			uint32(len(vertexArrayInfo.BufferBindings)),
		)
		binding := opengl.NewVertexArrayBufferBinding(
			a.VertexBuffer,
			0, data.Layout.TexCoordStride,
		)
		vertexArrayInfo.Attributes = append(vertexArrayInfo.Attributes, attribute)
		vertexArrayInfo.BufferBindings = append(vertexArrayInfo.BufferBindings, binding)
	}
	if data.Layout.HasColor {
		attribute := opengl.NewVertexArrayAttribute(
			ColorAttributeIndex,
			4, gl.FLOAT, false,
			uint32(data.Layout.ColorOffset),
			uint32(len(vertexArrayInfo.BufferBindings)),
		)
		binding := opengl.NewVertexArrayBufferBinding(
			a.VertexBuffer,
			0, data.Layout.ColorStride,
		)
		vertexArrayInfo.Attributes = append(vertexArrayInfo.Attributes, attribute)
		vertexArrayInfo.BufferBindings = append(vertexArrayInfo.BufferBindings, binding)
	}
	a.VertexArray.Allocate(vertexArrayInfo)
	return nil
}

func (a *VertexArray) Release() error {
	a.VertexArray.Release()
	a.IndexBuffer.Release()
	a.VertexBuffer.Release()
	return nil
}
