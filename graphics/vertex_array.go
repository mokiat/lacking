package graphics

import "github.com/go-gl/gl/v4.1-core/gl"

const (
	CoordAttributeIndex    = 0
	NormalAttributeIndex   = 1
	TangentAttributeIndex  = 2
	TexCoordAttributeIndex = 3
	ColorAttributeIndex    = 4
)

type VertexArray struct {
	ID             uint32
	VertexBufferID uint32
	IndexBufferID  uint32
}

type VertexArrayData struct {
	VertexData []byte
	IndexData  []byte
	Layout     VertexArrayLayout
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

func (a *VertexArray) Allocate(data VertexArrayData) error {
	gl.GenVertexArrays(1, &a.ID)
	gl.BindVertexArray(a.ID)

	gl.GenBuffers(1, &a.VertexBufferID)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.VertexBufferID)
	gl.BufferData(gl.ARRAY_BUFFER, len(data.VertexData), gl.Ptr(data.VertexData), gl.DYNAMIC_DRAW)

	if data.Layout.HasCoord {
		gl.EnableVertexAttribArray(CoordAttributeIndex)
		gl.VertexAttribPointer(CoordAttributeIndex, 3, gl.FLOAT, false, data.Layout.CoordStride, gl.PtrOffset(data.Layout.CoordOffset))
	}
	if data.Layout.HasNormal {
		gl.EnableVertexAttribArray(NormalAttributeIndex)
		gl.VertexAttribPointer(NormalAttributeIndex, 3, gl.FLOAT, false, data.Layout.NormalStride, gl.PtrOffset(data.Layout.NormalOffset))
	}
	if data.Layout.HasTangent {
		gl.EnableVertexAttribArray(TangentAttributeIndex)
		gl.VertexAttribPointer(TangentAttributeIndex, 3, gl.FLOAT, false, data.Layout.TangentStride, gl.PtrOffset(data.Layout.TangentOffset))
	}
	if data.Layout.HasTexCoord {
		gl.EnableVertexAttribArray(TexCoordAttributeIndex)
		gl.VertexAttribPointer(TexCoordAttributeIndex, 2, gl.FLOAT, false, data.Layout.TexCoordStride, gl.PtrOffset(data.Layout.TexCoordOffset))
	}
	if data.Layout.HasColor {
		gl.EnableVertexAttribArray(ColorAttributeIndex)
		gl.VertexAttribPointer(ColorAttributeIndex, 4, gl.FLOAT, false, data.Layout.ColorStride, gl.PtrOffset(data.Layout.ColorOffset))
	}

	gl.GenBuffers(1, &a.IndexBufferID)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, a.IndexBufferID)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(data.IndexData), gl.Ptr(data.IndexData), gl.STATIC_DRAW)
	return nil
}

func (a *VertexArray) Update(data VertexArrayData) error {
	gl.BindBuffer(gl.ARRAY_BUFFER, a.VertexBufferID)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(data.VertexData), gl.Ptr(data.VertexData))
	return nil
}

func (a *VertexArray) Release() error {
	gl.DeleteBuffers(1, &a.IndexBufferID)
	gl.DeleteBuffers(1, &a.VertexBufferID)
	gl.DeleteVertexArrays(1, &a.ID)
	a.ID = 0
	a.VertexBufferID = 0
	a.IndexBufferID = 0
	return nil
}
