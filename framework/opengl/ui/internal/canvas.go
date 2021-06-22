package internal

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

const maxVertexCount = 2048

func NewCanvas() *Canvas {
	return &Canvas{
		defaultLayer: &Layer{
			Translation: ui.NewPosition(0, 0),
			ClipBounds:  ui.NewBounds(0, 0, 1, 1),
			SolidColor:  ui.White(),
			StrokeColor: ui.Black(),
			StrokeSize:  1,
			Font:        nil,
		},
		topLayer: &Layer{},

		mesh: NewMesh(maxVertexCount),

		opaqueMaterial: NewDrawMaterial(),

		whiteMask: opengl.NewTwoDTexture(),
	}
}

var _ ui.Canvas = (*Canvas)(nil)

type Canvas struct {
	defaultLayer *Layer
	topLayer     *Layer
	currentLayer *Layer

	windowSize ui.Size

	mesh      *Mesh
	subMeshes []SubMesh

	opaqueMaterial *Material

	whiteMask *opengl.TwoDTexture
}

func (c *Canvas) Create() {
	c.mesh.Allocate()
	c.opaqueMaterial.Allocate()
	c.whiteMask.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             1,
		Height:            1,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
		Data:              []byte{0xFF, 0xFF, 0xFF, 0xFF},
	})
}

func (c *Canvas) Destroy() {
	c.whiteMask.Release()
	c.opaqueMaterial.Release()
	c.mesh.Release()
}

func (c *Canvas) Resize(width, height int) {
	c.windowSize = ui.NewSize(width, height)
	c.defaultLayer.ClipBounds.Size = c.windowSize
}

func (c *Canvas) ResizeFramebuffer(width, height int) {
	// TODO: Use own framebuffer which would allow for
	// only dirty region rerendering even when overlay.
}

func (c *Canvas) Begin() {
	c.currentLayer = c.topLayer
	c.mesh.Reset()
	c.subMeshes = c.subMeshes[:0]
}

func (c *Canvas) End() {
	c.mesh.Update()

	projectionMatrix := sprec.OrthoMat4(
		0.0, float32(c.windowSize.Width),
		0.0, float32(c.windowSize.Height),
		0.0, 1.0,
	).ColumnMajorArray()

	gl.Viewport(0, 0, int32(c.windowSize.Width), int32(c.windowSize.Height))
	gl.Enable(gl.FRAMEBUFFER_SRGB)
	gl.Clear(gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// TODO: Maybe optimize by accumulating draw commands
	// if they are similar.
	for _, subMesh := range c.subMeshes {
		material := subMesh.material
		gl.UseProgram(material.program.ID())
		gl.UniformMatrix4fv(material.projectionMatrixLocation, 1, false, &projectionMatrix[0])
		gl.BindTextureUnit(0, subMesh.texture.ID())
		gl.Uniform1i(material.textureLocation, 0)
		gl.BindVertexArray(c.mesh.vertexArray.ID())
		gl.DrawArrays(subMesh.primitive, int32(subMesh.vertexOffset), int32(subMesh.vertexCount))
	}

	// TODO: Remove once the remaining part of the framework
	// can handle resetting its settings.
	gl.Disable(gl.BLEND)
}

func (c *Canvas) Push() {
	c.currentLayer = c.currentLayer.Next()
}

func (c *Canvas) Pop() {
	c.currentLayer = c.currentLayer.Previous()
}

func (c *Canvas) Translate(delta ui.Position) {
	c.currentLayer.Translation = c.currentLayer.Translation.Translate(delta.X, delta.Y)
}

func (c *Canvas) Clip(bounds ui.Bounds) {
	c.currentLayer.ClipBounds = bounds.Translate(c.currentLayer.Translation)
}

func (c *Canvas) SolidColor() ui.Color {
	return c.currentLayer.SolidColor
}

func (c *Canvas) SetSolidColor(color ui.Color) {
	c.currentLayer.SolidColor = color
}

func (c *Canvas) StrokeColor() ui.Color {
	return c.currentLayer.StrokeColor
}

func (c *Canvas) SetStrokeColor(color ui.Color) {
	c.currentLayer.StrokeColor = color
}

func (c *Canvas) StrokeSize() int {
	return c.currentLayer.StrokeSize
}

func (c *Canvas) SetStrokeSize(size int) {
	c.currentLayer.StrokeSize = size
}

func (c *Canvas) Font() ui.Font {
	return c.currentLayer.Font
}

func (c *Canvas) SetFont(font ui.Font) {
	c.currentLayer.Font = font.(*Font)
}

func (c *Canvas) DrawRectangle(position ui.Position, size ui.Size) {
	// TODO
}

func (c *Canvas) FillRectangle(position ui.Position, size ui.Size) {
	color := c.currentLayer.SolidColor
	translation := sprec.NewVec2(
		float32(c.currentLayer.Translation.X),
		float32(c.currentLayer.Translation.Y),
	)

	vertTopLeft := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X),
			float32(position.Y),
		), translation),
		texCoord: sprec.NewVec2(0.0, 0.0),
		color:    color,
	}
	vertTopRight := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X+size.Width),
			float32(position.Y),
		), translation),
		texCoord: sprec.NewVec2(1.0, 0.0),
		color:    color,
	}
	vertBottomLeft := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X),
			float32(position.Y+size.Height),
		), translation),
		texCoord: sprec.NewVec2(0.0, 1.0),
		color:    color,
	}
	vertBottomRight := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X+size.Width),
			float32(position.Y+size.Height),
		), translation),
		texCoord: sprec.NewVec2(1.0, 1.0),
		color:    color,
	}

	offset := c.mesh.Offset()
	c.mesh.Append(vertTopLeft)
	c.mesh.Append(vertBottomLeft)
	c.mesh.Append(vertBottomRight)
	c.mesh.Append(vertTopLeft)
	c.mesh.Append(vertBottomRight)
	c.mesh.Append(vertTopRight)
	count := c.mesh.Offset() - offset

	c.subMeshes = append(c.subMeshes, SubMesh{
		material:     c.opaqueMaterial,
		texture:      c.whiteMask,
		vertexOffset: offset,
		vertexCount:  count,
		primitive:    gl.TRIANGLES,
	})
}

func (c *Canvas) DrawRoundRectangle(position ui.Position, size ui.Size, radius int) {
	// TODO
}

func (c *Canvas) FillRoundRectangle(position ui.Position, size ui.Size, radius int) {
	// TODO
}

func (c *Canvas) DrawCircle(position ui.Position, radius int) {
	// TODO
}

func (c *Canvas) FillCircle(position ui.Position, radius int) {
	// TODO
}

func (c *Canvas) DrawTriangle(first, second, third ui.Position) {
	// TODO
}

func (c *Canvas) FillTriangle(first, second, third ui.Position) {
	// TODO
}

func (c *Canvas) DrawLine(start, end ui.Position) {
	// TODO
}

func (c *Canvas) DrawImage(img ui.Image, position ui.Position, size ui.Size) {
	image := img.(*Image)

	translation := sprec.NewVec2(
		float32(c.currentLayer.Translation.X),
		float32(c.currentLayer.Translation.Y),
	)

	vertTopLeft := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X),
			float32(position.Y),
		), translation),
		texCoord: sprec.NewVec2(0.0, 0.0),
		color:    ui.White(),
	}
	vertTopRight := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X+size.Width),
			float32(position.Y),
		), translation),
		texCoord: sprec.NewVec2(1.0, 0.0),
		color:    ui.White(),
	}
	vertBottomLeft := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X),
			float32(position.Y+size.Height),
		), translation),
		texCoord: sprec.NewVec2(0.0, 1.0),
		color:    ui.White(),
	}
	vertBottomRight := Vertex{
		position: sprec.Vec2Sum(sprec.NewVec2(
			float32(position.X+size.Width),
			float32(position.Y+size.Height),
		), translation),
		texCoord: sprec.NewVec2(1.0, 1.0),
		color:    ui.White(),
	}

	offset := c.mesh.Offset()
	c.mesh.Append(vertTopLeft)
	c.mesh.Append(vertBottomLeft)
	c.mesh.Append(vertBottomRight)
	c.mesh.Append(vertTopLeft)
	c.mesh.Append(vertBottomRight)
	c.mesh.Append(vertTopRight)
	count := c.mesh.Offset() - offset

	c.subMeshes = append(c.subMeshes, SubMesh{
		material:     c.opaqueMaterial,
		texture:      image.texture,
		vertexOffset: offset,
		vertexCount:  count,
		primitive:    gl.TRIANGLES,
	})
}

func (c *Canvas) DrawText(text string, position ui.Position) {
	// TODO
}
