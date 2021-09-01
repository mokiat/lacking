package internal

import (
	"log"

	"github.com/go-gl/gl/v4.6-core/gl"
	"golang.org/x/image/font/sfnt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

const maxVertexCount = 2048

type Point struct {
	sprec.Vec2
	InStroke  ui.Stroke
	OutStroke ui.Stroke
}

type Shape struct {
	ui.Fill
	Points []Point
}

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

		mesh:        NewMesh(maxVertexCount),
		activeShape: &Shape{},

		opaqueMaterial:     NewDrawMaterial(),
		opaqueTessMaterial: NewPatchDrawMaterial(),

		whiteMask: opengl.NewTwoDTexture(),
	}
}

var _ ui.Canvas = (*Canvas)(nil)

type Canvas struct {
	defaultLayer *Layer
	topLayer     *Layer
	currentLayer *Layer

	windowSize ui.Size

	mesh        *Mesh
	subMeshes   []SubMesh
	activeShape *Shape

	opaqueMaterial     *Material
	opaqueTessMaterial *Material

	whiteMask *opengl.TwoDTexture
}

func (c *Canvas) Create() {
	c.mesh.Allocate()
	c.opaqueMaterial.Allocate()
	c.opaqueTessMaterial.Allocate()
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
	c.opaqueTessMaterial.Release()
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

	// gl.Enable(gl.CLIP_DISTANCE0)
	// gl.Enable(gl.CLIP_DISTANCE1)
	// gl.Enable(gl.CLIP_DISTANCE2)
	// gl.Enable(gl.CLIP_DISTANCE3)

	gl.Viewport(0, 0, int32(c.windowSize.Width), int32(c.windowSize.Height))
	gl.Enable(gl.FRAMEBUFFER_SRGB)
	gl.ClearStencil(0)
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
		if subMesh.skipColor {
			gl.ColorMask(false, false, false, false)
		} else {
			gl.ColorMask(true, true, true, true)
		}
		if subMesh.stencil {
			gl.Enable(gl.STENCIL_TEST)

			cfg := subMesh.stencilCfg
			gl.StencilFuncSeparate(gl.FRONT, cfg.stencilFuncFront.fn, cfg.stencilFuncFront.ref, cfg.stencilFuncFront.mask)
			gl.StencilFuncSeparate(gl.BACK, cfg.stencilFuncBack.fn, cfg.stencilFuncBack.ref, cfg.stencilFuncBack.mask)
			gl.StencilOpSeparate(gl.FRONT, cfg.stencilOpFront.sfail, cfg.stencilOpFront.dpfail, cfg.stencilOpFront.dppass)
			gl.StencilOpSeparate(gl.BACK, cfg.stencilOpBack.sfail, cfg.stencilOpBack.dpfail, cfg.stencilOpBack.dppass)
		} else {
			gl.Disable(gl.STENCIL_TEST)
		}
		if subMesh.culling {
			gl.Enable(gl.CULL_FACE)
		} else {
			gl.Disable(gl.CULL_FACE)
		}
		gl.CullFace(subMesh.cullFace)
		gl.UseProgram(material.program.ID())
		gl.UniformMatrix4fv(material.projectionMatrixLocation, 1, false, &projectionMatrix[0])
		gl.Uniform4f(material.clipDistancesLocation, subMesh.clipBounds.X, subMesh.clipBounds.Y, subMesh.clipBounds.Z, subMesh.clipBounds.W)
		gl.BindTextureUnit(0, subMesh.texture.ID())
		gl.Uniform1i(material.textureLocation, 0)
		gl.BindVertexArray(c.mesh.vertexArray.ID())
		if subMesh.patchVertices > 0 {
			gl.PatchParameteri(gl.PATCH_VERTICES, int32(subMesh.patchVertices))
		}
		gl.DrawArrays(subMesh.primitive, int32(subMesh.vertexOffset), int32(subMesh.vertexCount))
	}

	gl.ColorMask(true, true, true, true)
	gl.Disable(gl.STENCIL_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	// TODO: Remove once the remaining part of the framework
	// can handle resetting its settings.
	gl.Disable(gl.BLEND)

	// gl.Disable(gl.CLIP_DISTANCE0)
	// gl.Disable(gl.CLIP_DISTANCE1)
	// gl.Disable(gl.CLIP_DISTANCE2)
	// gl.Disable(gl.CLIP_DISTANCE3)
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
	// c.activeShape.SolidColor = color
}

func (c *Canvas) StrokeColor() ui.Color {
	return c.currentLayer.StrokeColor
}

func (c *Canvas) SetStrokeColor(color ui.Color) {
	c.currentLayer.StrokeColor = color
	// c.activeShape.StrokeColor = color
}

func (c *Canvas) StrokeSize() int {
	return c.currentLayer.StrokeSize
}

func (c *Canvas) SetStrokeSize(size int) {
	c.currentLayer.StrokeSize = size
	// c.activeShape.StrokeSize = size
}

func (c *Canvas) Font() ui.Font {
	return c.currentLayer.Font
}

func (c *Canvas) SetFont(font ui.Font) {
	c.currentLayer.Font = font.(*Font)
}

func (c *Canvas) FontSize() int {
	return int(c.currentLayer.FontSize)
}

func (c *Canvas) SetFontSize(size int) {
	c.currentLayer.FontSize = float32(size)
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
		clipBounds: sprec.NewVec4(
			float32(c.currentLayer.ClipBounds.X),
			float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
			float32(c.currentLayer.ClipBounds.Y),
			float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
		),
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
		clipBounds: sprec.NewVec4(
			float32(c.currentLayer.ClipBounds.X),
			float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
			float32(c.currentLayer.ClipBounds.Y),
			float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
		),
		material:     c.opaqueMaterial,
		texture:      image.texture,
		vertexOffset: offset,
		vertexCount:  count,
		primitive:    gl.TRIANGLES,
	})
}

func (c *Canvas) DrawText(text string, position ui.Position) {
	font := c.currentLayer.Font

	translation := sprec.Vec2Sum(
		sprec.NewVec2(
			float32(c.currentLayer.Translation.X),
			float32(c.currentLayer.Translation.Y),
		),
		sprec.NewVec2(
			float32(position.X),
			float32(position.Y),
		),
	)

	lastGlyph := (*fontGlyph)(nil)
	// offset := c.mesh.Offset()
	for _, ch := range text {
		log.Printf("%c\n", ch)

		lineHeight := font.lineHeight * c.currentLayer.FontSize
		lineAscent := font.lineAscent * c.currentLayer.FontSize
		if ch == '\r' {
			translation.X = float32(c.currentLayer.Translation.X + position.X)
			lastGlyph = nil
			continue
		}
		if ch == '\n' {
			translation.X = float32(c.currentLayer.Translation.X + position.X)
			translation.Y += lineHeight
			lastGlyph = nil
			continue
		}

		const scale = 1

		glyph := font.glyphs[ch]
		advance := glyph.advance * c.currentLayer.FontSize
		leftBearing := glyph.leftBearing * c.currentLayer.FontSize
		// rightBearing := glyph.rightBearing * c.currentLayer.FontSize
		ascent := glyph.ascent * c.currentLayer.FontSize
		// descent := glyph.descent * c.currentLayer.FontSize

		stroke := ui.Stroke{
			Size: 0,
		}
		c.Push()
		c.Translate(ui.NewPosition(
			int(translation.X),
			int(translation.Y),
		))
		c.Translate(ui.NewPosition(
			int(leftBearing)*scale,
			int(lineAscent-ascent)*scale,
		))

		log.Println("begin shape...")
		c.BeginShape(ui.Fill{
			Winding:         ui.WindingCW, // non-standard
			BackgroundColor: ui.Red(),
			Rule:            ui.FillRuleEvenOdd,
		})
		for _, segment := range glyph.segments {
			switch segment.Op {
			case sfnt.SegmentOpMoveTo:
				log.Printf("move to (%d, %d)\n",
					segment.Args[0].X.Floor()*scale,
					segment.Args[0].Y.Floor()*scale+400,
				)
				c.MoveTo(
					ui.NewPosition(
						segment.Args[0].X.Floor()*scale,
						segment.Args[0].Y.Floor()*scale,
					),
				)

			case sfnt.SegmentOpLineTo:
				log.Printf("line to (%d, %d)\n",
					segment.Args[0].X.Floor()*scale,
					segment.Args[0].Y.Floor()*scale+400,
				)
				c.LineTo(
					ui.NewPosition(
						segment.Args[0].X.Floor()*scale,
						segment.Args[0].Y.Floor()*scale,
					),
					stroke, stroke,
				)

			case sfnt.SegmentOpQuadTo:
				log.Printf("quad to (%d, %d) (%d, %d)\n",
					segment.Args[0].X.Floor()*scale,
					segment.Args[0].Y.Floor()*scale+400,
					segment.Args[1].X.Floor()*scale,
					segment.Args[1].Y.Floor()*scale+400,
				)
				c.QuadTo(
					ui.NewPosition(
						segment.Args[0].X.Floor()*scale,
						segment.Args[0].Y.Floor()*scale,
					),
					ui.NewPosition(
						segment.Args[1].X.Floor()*scale,
						segment.Args[1].Y.Floor()*scale,
					),
					stroke, stroke,
				)

			case sfnt.SegmentOpCubeTo:
				log.Printf("cube to (%d, %d) (%d, %d) (%d, %d)\n",
					segment.Args[0].X.Floor()*scale,
					segment.Args[0].Y.Floor()*scale+400,
					segment.Args[1].X.Floor()*scale,
					segment.Args[1].Y.Floor()*scale+400,
					segment.Args[2].X.Floor()*scale,
					segment.Args[2].Y.Floor()*scale+400,
				)
				c.CubeTo(
					ui.NewPosition(
						segment.Args[0].X.Floor()*scale,
						segment.Args[0].Y.Floor()*scale,
					),
					ui.NewPosition(
						segment.Args[1].X.Floor()*scale,
						segment.Args[1].Y.Floor()*scale,
					),
					ui.NewPosition(
						segment.Args[2].X.Floor()*scale,
						segment.Args[2].Y.Floor()*scale,
					),
					stroke, stroke,
				)
			default:
				log.Println("unknown")
			}
		}
		log.Println("end shape...")
		c.EndShape()
		c.Pop()

		// vertTopLeft := Vertex{
		// 	position: sprec.Vec2Sum(sprec.NewVec2(
		// 		leftBearing,
		// 		lineAscent-ascent,
		// 	), translation),
		// 	texCoord: sprec.NewVec2(glyph.leftU, glyph.topV),
		// 	color:    c.currentLayer.SolidColor,
		// }
		// vertTopRight := Vertex{
		// 	position: sprec.Vec2Sum(sprec.NewVec2(
		// 		advance-rightBearing,
		// 		lineAscent-ascent,
		// 	), translation),
		// 	texCoord: sprec.NewVec2(glyph.rightU, glyph.topV),
		// 	color:    c.currentLayer.SolidColor,
		// }
		// vertBottomLeft := Vertex{
		// 	position: sprec.Vec2Sum(sprec.NewVec2(
		// 		leftBearing,
		// 		lineAscent+descent,
		// 	), translation),
		// 	texCoord: sprec.NewVec2(glyph.leftU, glyph.bottomV),
		// 	color:    c.currentLayer.SolidColor,
		// }
		// vertBottomRight := Vertex{
		// 	position: sprec.Vec2Sum(sprec.NewVec2(
		// 		advance-rightBearing,
		// 		lineAscent+descent,
		// 	), translation),
		// 	texCoord: sprec.NewVec2(glyph.rightU, glyph.bottomV),
		// 	color:    c.currentLayer.SolidColor,
		// }

		// c.mesh.Append(vertTopLeft)
		// c.mesh.Append(vertBottomLeft)
		// c.mesh.Append(vertBottomRight)
		// c.mesh.Append(vertTopLeft)
		// c.mesh.Append(vertBottomRight)
		// c.mesh.Append(vertTopRight)

		translation.X += advance * scale * 5
		if lastGlyph != nil {
			translation.X += lastGlyph.kerns[ch] * c.currentLayer.FontSize
		}
		lastGlyph = glyph

		// if true {
		// 	break
		// }
	}
	// count := c.mesh.Offset() - offset

	// c.subMeshes = append(c.subMeshes, SubMesh{
	// 	clipBounds: sprec.NewVec4(
	// 		float32(c.currentLayer.ClipBounds.X),
	// 		float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
	// 		float32(c.currentLayer.ClipBounds.Y),
	// 		float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
	// 	),
	// 	material:     c.opaqueMaterial,
	// 	texture:      font.texture,
	// 	vertexOffset: offset,
	// 	vertexCount:  count,
	// 	primitive:    gl.TRIANGLES,
	// })
}

func (c *Canvas) TextSize(text string) ui.Size {
	font := c.currentLayer.Font
	fontSize := c.currentLayer.FontSize

	if len(text) == 0 {
		return ui.NewSize(0, 0)
	}

	result := sprec.NewVec2(0, font.lineAscent+font.lineDescent)
	currentWidth := float32(0.0)
	lastGlyph := (*fontGlyph)(nil)
	for _, ch := range text {
		if ch == '\r' {
			result.X = sprec.Max(result.X, currentWidth)
			currentWidth = 0.0
			lastGlyph = nil
			continue
		}
		if ch == '\n' {
			result.X = sprec.Max(result.X, currentWidth)
			result.Y += font.lineHeight - (font.lineAscent + font.lineDescent)
			currentWidth = 0.0
			lastGlyph = nil
			continue
		}
		glyph := font.glyphs[ch]
		currentWidth += glyph.advance
		if lastGlyph != nil {
			currentWidth += lastGlyph.kerns[ch]
		}
		lastGlyph = glyph
	}
	result.X = sprec.Max(result.X, currentWidth)
	result = sprec.Vec2Prod(result, fontSize)

	return ui.NewSize(int(result.X), int(result.Y))
}

func (c *Canvas) BeginShape(fill ui.Fill) {
	c.activeShape.Fill = fill
	c.activeShape.Points = c.activeShape.Points[:0]
}

func (c *Canvas) MoveTo(position ui.Position) {
	c.activeShape.Points = append(c.activeShape.Points, Point{
		Vec2: sprec.NewVec2(float32(position.X), float32(position.Y)),
	})
}

func (c *Canvas) LineTo(position ui.Position, startStroke, endStroke ui.Stroke) {
	c.activeShape.Points[len(c.activeShape.Points)-1].OutStroke = startStroke
	c.activeShape.Points = append(c.activeShape.Points, Point{
		Vec2:     sprec.NewVec2(float32(position.X), float32(position.Y)),
		InStroke: endStroke,
	})
}

func (c *Canvas) QuadTo(control, position ui.Position, startStroke, endStroke ui.Stroke) {
	startPoint := c.activeShape.Points[len(c.activeShape.Points)-1]
	startPoint.OutStroke = startStroke

	sX := startPoint.X
	sY := startPoint.Y
	cX := float32(control.X)
	cY := float32(control.Y)
	eX := float32(position.X)
	eY := float32(position.Y)

	const precision = 30 // TODO: Evaluate based on points
	for i := 1; i <= precision; i++ {
		// TODO: Use derivatives for performance improvement
		t := float32(i) / float32(precision)
		alpha := (1 - t) * (1 - t)
		beta := t * t
		c.activeShape.Points = append(c.activeShape.Points, Point{
			Vec2: sprec.NewVec2(
				cX+alpha*(sX-cX)+beta*(eX-cX),
				cY+alpha*(sY-cY)+beta*(eY-cY),
			),
			InStroke:  endStroke, // TODO: Interpolate
			OutStroke: endStroke, // TODO: Interpolate and set only if non-last
		})
	}
}

func (c *Canvas) CubeTo(control1, control2, position ui.Position, startStroke, endStroke ui.Stroke) {
	startPoint := c.activeShape.Points[len(c.activeShape.Points)-1]
	startPoint.OutStroke = startStroke

	sX := startPoint.X
	sY := startPoint.Y
	c1X := float32(control1.X)
	c1Y := float32(control1.Y)
	c2X := float32(control2.X)
	c2Y := float32(control2.Y)
	eX := float32(position.X)
	eY := float32(position.Y)

	const precision = 30 // TODO: Evaluate based on points
	for i := 1; i <= precision; i++ {
		// TODO: Use derivatives for performance improvement
		t := float32(i) / float32(precision)
		alpha := (1 - t) * (1 - t) * (1 - t)
		beta := 3 * (1 - t) * (1 - t) * t
		gamma := 3 * (1 - t) * t * t
		delta := t * t * t
		c.activeShape.Points = append(c.activeShape.Points, Point{
			Vec2: sprec.NewVec2(
				alpha*sX+beta*c1X+gamma*c2X+delta*eX,
				alpha*sY+beta*c1Y+gamma*c2Y+delta*eY,
			),
			InStroke:  endStroke, // TODO: Interpolate
			OutStroke: endStroke, // TODO: Interpolate and set only if non-last
		})
	}
}

func (c *Canvas) CloseLoop(startStroke, endStroke ui.Stroke) {
	c.activeShape.Points[len(c.activeShape.Points)-1].OutStroke = startStroke
	c.activeShape.Points = append(c.activeShape.Points, Point{
		Vec2:     c.activeShape.Points[0].Vec2,
		InStroke: endStroke,
	})
}

// func (c *Canvas) EndShape() {
// 	translation := sprec.NewVec2(
// 		float32(c.currentLayer.Translation.X),
// 		float32(c.currentLayer.Translation.Y),
// 	)

// 	offset := c.mesh.Offset()
// 	// TODO: Only if background color or texture is set
// 	for _, segment := range c.activeShape.Segments {
// 		c.mesh.Append(Vertex{
// 			position: sprec.Vec2Sum(segment.A.position, translation),
// 			texCoord: sprec.NewVec2(0.0, 0.0),
// 			color:    c.activeShape.Fill.BackgroundColor,
// 		})
// 		c.mesh.Append(Vertex{
// 			position: sprec.Vec2Sum(segment.B.position, translation),
// 			texCoord: sprec.NewVec2(0.0, 0.0),
// 			color:    c.activeShape.Fill.BackgroundColor,
// 		})
// 		c.mesh.Append(Vertex{
// 			position: sprec.Vec2Sum(segment.C.position, translation),
// 			texCoord: sprec.NewVec2(0.0, 0.0),
// 			color:    c.activeShape.Fill.BackgroundColor,
// 		})
// 		c.mesh.Append(Vertex{
// 			position: sprec.Vec2Sum(segment.CP1.position, translation),
// 			texCoord: sprec.NewVec2(0.0, 0.0),
// 			color:    c.activeShape.Fill.BackgroundColor,
// 		})
// 		c.mesh.Append(Vertex{
// 			position: sprec.Vec2Sum(segment.CP2.position, translation),
// 			texCoord: sprec.NewVec2(0.0, 0.0),
// 			color:    c.activeShape.Fill.BackgroundColor,
// 		})
// 	}
// 	count := c.mesh.Offset() - offset

// 	c.subMeshes = append(c.subMeshes, SubMesh{
// 		clipBounds: sprec.NewVec4(
// 			float32(c.currentLayer.ClipBounds.X),
// 			float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
// 			float32(c.currentLayer.ClipBounds.Y),
// 			float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
// 		),
// 		material:      c.opaqueTessMaterial,
// 		texture:       c.whiteMask,
// 		vertexOffset:  offset,
// 		vertexCount:   count,
// 		patchVertices: 5,
// 		primitive:     gl.PATCHES,
// 	})

// 	// TODO: Draw stroke
// }

func (c *Canvas) EndShape() {
	translation := sprec.NewVec2(
		float32(c.currentLayer.Translation.X),
		float32(c.currentLayer.Translation.Y),
	)

	offset := c.mesh.Offset()
	// TODO: Only if background color or texture is set
	for _, point := range c.activeShape.Points {
		c.mesh.Append(Vertex{
			position: sprec.Vec2Sum(point.Vec2, translation),
			texCoord: sprec.NewVec2(0.0, 0.0),
			color:    c.activeShape.BackgroundColor,
		})
	}
	count := c.mesh.Offset() - offset

	cullFace := gl.BACK
	if c.activeShape.Winding == ui.WindingCW {
		cullFace = gl.FRONT
	}

	if c.activeShape.Rule == ui.FillRuleSimple {
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
		})
	}

	if c.activeShape.Rule != ui.FillRuleSimple {
		// clear stencil
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			culling:      false,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
			skipColor:    true,
			stencil:      true,
			stencilCfg: stencilConfig{
				stencilFuncFront: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilFuncBack: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilOpFront: stencilOp{
					sfail:  gl.REPLACE,
					dpfail: gl.REPLACE,
					dppass: gl.REPLACE,
				},
				stencilOpBack: stencilOp{
					sfail:  gl.REPLACE,
					dpfail: gl.REPLACE,
					dppass: gl.REPLACE,
				},
			},
		})

		// render stencil mask
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
			skipColor:    true, // we don't want to render anything
			stencil:      true,
			stencilCfg: stencilConfig{
				stencilFuncFront: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilFuncBack: stencilFunc{
					fn:   gl.ALWAYS,
					ref:  0,
					mask: 0xFF,
				},
				stencilOpFront: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.INCR_WRAP, // increase correct winding
				},
				stencilOpBack: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.DECR_WRAP, // decrease incorrect winding
				},
			},
		})

		// render final polygon
		c.subMeshes = append(c.subMeshes, SubMesh{
			clipBounds: sprec.NewVec4(
				float32(c.currentLayer.ClipBounds.X),
				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
				float32(c.currentLayer.ClipBounds.Y),
				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
			),
			material:     c.opaqueMaterial,
			texture:      c.whiteMask,
			vertexOffset: offset,
			vertexCount:  count,
			cullFace:     uint32(cullFace),
			primitive:    gl.TRIANGLE_FAN,
			skipColor:    false, // we want to render now
			stencil:      true,
			stencilCfg: stencilConfig{
				stencilFuncFront: stencilFunc{
					fn:   gl.LESS,
					ref:  0,
					mask: 0xFF,
				},
				stencilFuncBack: stencilFunc{
					fn:   gl.LESS,
					ref:  0,
					mask: 0xFF,
				},
				stencilOpFront: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.KEEP,
				},
				stencilOpBack: stencilOp{
					sfail:  gl.KEEP,
					dpfail: gl.KEEP,
					dppass: gl.KEEP,
				},
			},
		})
	}

	// TODO: Draw stroke
}
