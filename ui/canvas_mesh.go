package ui

import (
	"encoding/binary"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/render"
)

const (
	shapePositionAttribIndex = 0
	shapeMeshVertexSize      = 2 * 4

	contourPositionAttribIndex = 0
	contourColorAttribIndex    = 2
	contourMeshVertexSize      = 2*4 + 1*4

	textPositionAttribIndex = 0
	textTexCoordAttribIndex = 1
	textMeshVertexSize      = 2*4 + 2*4
)

func newShapeMesh(vertexCount int) *shapeMesh {
	data := make([]byte, vertexCount*shapeMeshVertexSize)
	return &shapeMesh{
		vertexData:    data,
		vertexPlotter: buffer.NewPlotter(data, binary.LittleEndian),
	}
}

type shapeMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  render.Buffer
	vertexArray   render.VertexArray
}

func (m *shapeMesh) Allocate(api render.API) {
	m.vertexBuffer = api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(m.vertexData),
	})

	m.vertexArray = api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: m.vertexBuffer,
				Stride:       shapeMeshVertexSize,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: shapePositionAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
		},
	})
}

func (m *shapeMesh) Release() {
	m.vertexArray.Release()
	m.vertexBuffer.Release()
}

func (m *shapeMesh) Update() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.vertexBuffer.Update(render.BufferUpdateInfo{
			Data:   m.vertexData[:length],
			Offset: 0,
		})
	}
}

func (m *shapeMesh) Reset() {
	m.vertexOffset = 0
	m.vertexPlotter.Rewind()
}

func (m *shapeMesh) Offset() int {
	return m.vertexOffset
}

func (m *shapeMesh) Append(vertex shapeVertex) {
	m.vertexPlotter.PlotFloat32(vertex.position.X)
	m.vertexPlotter.PlotFloat32(vertex.position.Y)
	m.vertexOffset++
}

type shapeVertex struct {
	position sprec.Vec2
}

func newContourMesh(vertexCount int) *contourMesh {
	data := make([]byte, vertexCount*contourMeshVertexSize)
	return &contourMesh{
		vertexData:    data,
		vertexPlotter: buffer.NewPlotter(data, binary.LittleEndian),
	}
}

type contourMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  render.Buffer
	vertexArray   render.VertexArray
}

func (m *contourMesh) Allocate(api render.API) {
	m.vertexBuffer = api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(m.vertexData),
	})

	m.vertexArray = api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: m.vertexBuffer,
				Stride:       contourMeshVertexSize,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: contourPositionAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
			{
				Binding:  0,
				Location: contourColorAttribIndex,
				Format:   render.VertexAttributeFormatRGBA8UN,
				Offset:   2 * 4,
			},
		},
	})
}

func (m *contourMesh) Release() {
	m.vertexArray.Release()
	m.vertexBuffer.Release()
}

func (m *contourMesh) Update() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.vertexBuffer.Update(render.BufferUpdateInfo{
			Data:   m.vertexData[:length],
			Offset: 0,
		})
	}
}

func (m *contourMesh) Reset() {
	m.vertexOffset = 0
	m.vertexPlotter.Rewind()
}

func (m *contourMesh) Offset() int {
	return m.vertexOffset
}

func (m *contourMesh) Append(vertex contourVertex) {
	m.vertexPlotter.PlotFloat32(vertex.position.X)
	m.vertexPlotter.PlotFloat32(vertex.position.Y)
	m.vertexPlotter.PlotByte(vertex.color.R)
	m.vertexPlotter.PlotByte(vertex.color.G)
	m.vertexPlotter.PlotByte(vertex.color.B)
	m.vertexPlotter.PlotByte(vertex.color.A)
	m.vertexOffset++
}

type contourVertex struct {
	position sprec.Vec2
	color    Color
}

func newTextMesh(vertexCount int) *textMesh {
	data := make([]byte, vertexCount*textMeshVertexSize)
	return &textMesh{
		vertexData:    data,
		vertexPlotter: buffer.NewPlotter(data, binary.LittleEndian),
	}
}

type textMesh struct {
	vertexData    []byte
	vertexPlotter *buffer.Plotter
	vertexOffset  int
	vertexBuffer  render.Buffer
	vertexArray   render.VertexArray
}

func (m *textMesh) Allocate(api render.API) {
	m.vertexBuffer = api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(m.vertexData),
	})

	m.vertexArray = api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: m.vertexBuffer,
				Stride:       textMeshVertexSize,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: textPositionAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   0,
			},
			{
				Binding:  0,
				Location: textTexCoordAttribIndex,
				Format:   render.VertexAttributeFormatRG32F,
				Offset:   2 * 4,
			},
		},
	})
}

func (m *textMesh) Release() {
	m.vertexArray.Release()
	m.vertexBuffer.Release()
}

func (m *textMesh) Update() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.vertexBuffer.Update(render.BufferUpdateInfo{
			Data:   m.vertexData[:length],
			Offset: 0,
		})
	}
}

func (m *textMesh) Reset() {
	m.vertexOffset = 0
	m.vertexPlotter.Rewind()
}

func (m *textMesh) Offset() int {
	return m.vertexOffset
}

func (m *textMesh) Append(vertex textVertex) {
	m.vertexPlotter.PlotFloat32(vertex.position.X)
	m.vertexPlotter.PlotFloat32(vertex.position.Y)
	m.vertexPlotter.PlotFloat32(vertex.texCoord.X)
	m.vertexPlotter.PlotFloat32(vertex.texCoord.Y)
	m.vertexOffset++
}

type textVertex struct {
	position sprec.Vec2
	texCoord sprec.Vec2
}
