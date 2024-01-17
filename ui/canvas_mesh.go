package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
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

func newShapeMesh(api render.API, vertexCount int) *shapeMesh {
	data := make([]byte, vertexCount*shapeMeshVertexSize)
	return &shapeMesh{
		api:           api,
		vertexData:    data,
		vertexPlotter: blob.NewPlotter(data),
	}
}

type shapeMesh struct {
	api           render.API
	vertexData    []byte
	vertexPlotter *blob.Plotter
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
		Bindings: []render.VertexArrayBinding{
			render.NewVertexArrayBinding(m.vertexBuffer, shapeMeshVertexSize),
		},
		Attributes: []render.VertexArrayAttribute{
			render.NewVertexArrayAttribute(0, shapePositionAttribIndex, 0, render.VertexAttributeFormatRG32F),
		},
	})
}

func (m *shapeMesh) Release() {
	defer m.vertexBuffer.Release()
	defer m.vertexArray.Release()
}

func (m *shapeMesh) Upload() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.api.Queue().WriteBuffer(m.vertexBuffer, 0, m.vertexData[:length])
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

func newContourMesh(api render.API, vertexCount int) *contourMesh {
	data := make([]byte, vertexCount*contourMeshVertexSize)
	return &contourMesh{
		api:           api,
		vertexData:    data,
		vertexPlotter: blob.NewPlotter(data),
	}
}

type contourMesh struct {
	api           render.API
	vertexData    []byte
	vertexPlotter *blob.Plotter
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
		Bindings: []render.VertexArrayBinding{
			render.NewVertexArrayBinding(m.vertexBuffer, contourMeshVertexSize),
		},
		Attributes: []render.VertexArrayAttribute{
			render.NewVertexArrayAttribute(0, contourPositionAttribIndex, 0, render.VertexAttributeFormatRG32F),
			render.NewVertexArrayAttribute(0, contourColorAttribIndex, 2*4, render.VertexAttributeFormatRGBA8UN),
		},
	})
}

func (m *contourMesh) Release() {
	defer m.vertexBuffer.Release()
	defer m.vertexArray.Release()
}

func (m *contourMesh) Upload() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.api.Queue().WriteBuffer(m.vertexBuffer, 0, m.vertexData[:length])
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
	m.vertexPlotter.PlotUint8(vertex.color.R)
	m.vertexPlotter.PlotUint8(vertex.color.G)
	m.vertexPlotter.PlotUint8(vertex.color.B)
	m.vertexPlotter.PlotUint8(vertex.color.A)
	m.vertexOffset++
}

type contourVertex struct {
	position sprec.Vec2
	color    Color
}

func newTextMesh(api render.API, vertexCount int) *textMesh {
	data := make([]byte, vertexCount*textMeshVertexSize)
	return &textMesh{
		api:           api,
		vertexData:    data,
		vertexPlotter: blob.NewPlotter(data),
	}
}

type textMesh struct {
	api           render.API
	vertexData    []byte
	vertexPlotter *blob.Plotter
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
		Bindings: []render.VertexArrayBinding{
			render.NewVertexArrayBinding(m.vertexBuffer, textMeshVertexSize),
		},
		Attributes: []render.VertexArrayAttribute{
			render.NewVertexArrayAttribute(0, textPositionAttribIndex, 0, render.VertexAttributeFormatRG32F),
			render.NewVertexArrayAttribute(0, textTexCoordAttribIndex, 2*4, render.VertexAttributeFormatRG32F),
		},
	})
}

func (m *textMesh) Release() {
	defer m.vertexBuffer.Release()
	defer m.vertexArray.Release()
}

func (m *textMesh) Upload() {
	if length := m.vertexPlotter.Offset(); length > 0 {
		m.api.Queue().WriteBuffer(m.vertexBuffer, 0, m.vertexData[:length])
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
