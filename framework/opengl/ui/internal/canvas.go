package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

func NewCanvas(renderer *Renderer) *Canvas {
	return &Canvas{
		renderer: renderer,
		defaultLayer: &Layer{
			Translation: ui.NewPosition(0, 0),
			ClipBounds:  ui.NewBounds(0, 0, 1, 1),
			SolidColor:  ui.White(),
			StrokeColor: ui.Black(),
			StrokeSize:  1,
			Font:        nil,
		},
		topLayer: &Layer{},
	}
}

var _ ui.Canvas = (*Canvas)(nil)

type Canvas struct {
	renderer *Renderer

	defaultLayer *Layer
	topLayer     *Layer
	currentLayer *Layer

	windowSize ui.Size
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

	c.renderer.Begin(Target{
		Framebuffer: opengl.DefaultFramebuffer(),
		Size: sprec.NewVec2(
			float32(c.windowSize.Width),
			float32(c.windowSize.Height),
		),
	})
}

func (c *Canvas) End() {
	c.renderer.End()
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
	if c.currentLayer.SolidColor.Transparent() {
		return
	}

	c.renderer.SetClipBounds(
		float32(c.currentLayer.ClipBounds.X),
		float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
		float32(c.currentLayer.ClipBounds.Y),
		float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
	)
	c.renderer.SetTransform(sprec.TranslationMat4(
		float32(c.currentLayer.Translation.X),
		float32(c.currentLayer.Translation.Y),
		0.0,
	))

	shape := c.renderer.BeginShape(Fill{
		color: sprec.NewVec4(
			float32(c.currentLayer.SolidColor.R)/255.0,
			float32(c.currentLayer.SolidColor.G)/255.0,
			float32(c.currentLayer.SolidColor.B)/255.0,
			float32(c.currentLayer.SolidColor.A)/255.0,
		),
	})
	shape.MoveTo(sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	))
	shape.LineTo(sprec.NewVec2(
		float32(position.X),
		float32(position.Y+size.Height),
	))
	shape.LineTo(sprec.NewVec2(
		float32(position.X+size.Width),
		float32(position.Y+size.Height),
	))
	shape.LineTo(sprec.NewVec2(
		float32(position.X+size.Width),
		float32(position.Y),
	))
	c.renderer.EndShape(shape)
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
	c.renderer.SetClipBounds(
		float32(c.currentLayer.ClipBounds.X),
		float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
		float32(c.currentLayer.ClipBounds.Y),
		float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
	)
	c.renderer.SetTransform(sprec.TranslationMat4(
		float32(c.currentLayer.Translation.X),
		float32(c.currentLayer.Translation.Y),
		0.0,
	))
	c.renderer.SetTextureTransform(sprec.Mat4MultiProd(
		sprec.ScaleMat4(
			1.0/float32(size.Width),
			1.0/float32(size.Height),
			1.0,
		),
		sprec.TranslationMat4(
			-float32(position.X),
			-float32(position.Y),
			0.0,
		),
	))

	shape := c.renderer.BeginShape(Fill{
		color: sprec.NewVec4(1.0, 1.0, 1.0, 1.0),
		image: img.(*Image),
	})
	shape.MoveTo(sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	))
	shape.LineTo(sprec.NewVec2(
		float32(position.X),
		float32(position.Y+size.Height),
	))
	shape.LineTo(sprec.NewVec2(
		float32(position.X+size.Width),
		float32(position.Y+size.Height),
	))
	shape.LineTo(sprec.NewVec2(
		float32(position.X+size.Width),
		float32(position.Y),
	))
	c.renderer.EndShape(shape)
}

func (c *Canvas) DrawText(text string, position ui.Position) {
	// font := c.currentLayer.Font

	// translation := sprec.Vec2Sum(
	// 	sprec.NewVec2(
	// 		float32(c.currentLayer.Translation.X),
	// 		float32(c.currentLayer.Translation.Y),
	// 	),
	// 	sprec.NewVec2(
	// 		float32(position.X),
	// 		float32(position.Y),
	// 	),
	// )

	// lastGlyph := (*fontGlyph)(nil)
	// offset := c.mesh.Offset()
	// for _, ch := range text {
	// 	lineHeight := font.lineHeight * c.currentLayer.FontSize
	// 	lineAscent := font.lineAscent * c.currentLayer.FontSize
	// 	if ch == '\r' {
	// 		translation.X = float32(c.currentLayer.Translation.X + position.X)
	// 		lastGlyph = nil
	// 		continue
	// 	}
	// 	if ch == '\n' {
	// 		translation.X = float32(c.currentLayer.Translation.X + position.X)
	// 		translation.Y += lineHeight
	// 		lastGlyph = nil
	// 		continue
	// 	}

	// 	if glyph, ok := font.glyphs[ch]; ok {

	// 		advance := glyph.advance * c.currentLayer.FontSize
	// 		leftBearing := glyph.leftBearing * c.currentLayer.FontSize
	// 		rightBearing := glyph.rightBearing * c.currentLayer.FontSize
	// 		ascent := glyph.ascent * c.currentLayer.FontSize
	// 		descent := glyph.descent * c.currentLayer.FontSize

	// 		color := sprec.NewVec4(
	// 			float32(c.currentLayer.SolidColor.R)/255.0,
	// 			float32(c.currentLayer.SolidColor.G)/255.0,
	// 			float32(c.currentLayer.SolidColor.B)/255.0,
	// 			float32(c.currentLayer.SolidColor.A)/255.0,
	// 		)

	// 		vertTopLeft := Vertex{
	// 			position: sprec.Vec2Sum(sprec.NewVec2(
	// 				leftBearing,
	// 				lineAscent-ascent,
	// 			), translation),
	// 			texCoord: sprec.NewVec2(glyph.leftU, glyph.topV),
	// 			color:    color,
	// 		}
	// 		vertTopRight := Vertex{
	// 			position: sprec.Vec2Sum(sprec.NewVec2(
	// 				advance-rightBearing,
	// 				lineAscent-ascent,
	// 			), translation),
	// 			texCoord: sprec.NewVec2(glyph.rightU, glyph.topV),
	// 			color:    color,
	// 		}
	// 		vertBottomLeft := Vertex{
	// 			position: sprec.Vec2Sum(sprec.NewVec2(
	// 				leftBearing,
	// 				lineAscent+descent,
	// 			), translation),
	// 			texCoord: sprec.NewVec2(glyph.leftU, glyph.bottomV),
	// 			color:    color,
	// 		}
	// 		vertBottomRight := Vertex{
	// 			position: sprec.Vec2Sum(sprec.NewVec2(
	// 				advance-rightBearing,
	// 				lineAscent+descent,
	// 			), translation),
	// 			texCoord: sprec.NewVec2(glyph.rightU, glyph.bottomV),
	// 			color:    color,
	// 		}

	// 		c.mesh.Append(vertTopLeft)
	// 		c.mesh.Append(vertBottomLeft)
	// 		c.mesh.Append(vertBottomRight)
	// 		c.mesh.Append(vertTopLeft)
	// 		c.mesh.Append(vertBottomRight)
	// 		c.mesh.Append(vertTopRight)

	// 		translation.X += advance
	// 		if lastGlyph != nil {
	// 			translation.X += lastGlyph.kerns[ch] * c.currentLayer.FontSize
	// 		}
	// 		lastGlyph = glyph
	// 	}
	// }
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
	// TODO: Move inside font?
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
		if glyph, ok := font.glyphs[ch]; ok {
			currentWidth += glyph.advance
			if lastGlyph != nil {
				currentWidth += lastGlyph.kerns[ch]
			}
			lastGlyph = glyph
		}
	}
	result.X = sprec.Max(result.X, currentWidth)
	result = sprec.Vec2Prod(result, fontSize)

	return ui.NewSize(int(result.X), int(result.Y))
}

// func (c *Canvas) BeginShape(fill ui.Fill) {
// 	c.activeShape.Fill = fill
// 	c.activeShape.Points = c.activeShape.Points[:0]
// }

// func (c *Canvas) MoveTo(position ui.Position) {
// 	c.activeShape.Points = append(c.activeShape.Points, Point{
// 		Vec2: sprec.NewVec2(float32(position.X), float32(position.Y)),
// 	})
// }

// func (c *Canvas) LineTo(position ui.Position, startStroke, endStroke ui.Stroke) {
// 	c.activeShape.Points[len(c.activeShape.Points)-1].OutStroke = startStroke
// 	c.activeShape.Points = append(c.activeShape.Points, Point{
// 		Vec2:     sprec.NewVec2(float32(position.X), float32(position.Y)),
// 		InStroke: endStroke,
// 	})
// }

// func (c *Canvas) QuadTo(control, position ui.Position, startStroke, endStroke ui.Stroke) {
// 	startPoint := c.activeShape.Points[len(c.activeShape.Points)-1]
// 	startPoint.OutStroke = startStroke

// 	sX := startPoint.X
// 	sY := startPoint.Y
// 	cX := float32(control.X)
// 	cY := float32(control.Y)
// 	eX := float32(position.X)
// 	eY := float32(position.Y)

// 	const precision = 30 // TODO: Evaluate based on points
// 	for i := 1; i <= precision; i++ {
// 		// TODO: Use derivatives for performance improvement
// 		t := float32(i) / float32(precision)
// 		alpha := (1 - t) * (1 - t)
// 		beta := t * t
// 		c.activeShape.Points = append(c.activeShape.Points, Point{
// 			Vec2: sprec.NewVec2(
// 				cX+alpha*(sX-cX)+beta*(eX-cX),
// 				cY+alpha*(sY-cY)+beta*(eY-cY),
// 			),
// 			InStroke:  endStroke, // TODO: Interpolate
// 			OutStroke: endStroke, // TODO: Interpolate and set only if non-last
// 		})
// 	}
// }

// func (c *Canvas) CubeTo(control1, control2, position ui.Position, startStroke, endStroke ui.Stroke) {
// 	startPoint := c.activeShape.Points[len(c.activeShape.Points)-1]
// 	startPoint.OutStroke = startStroke

// 	sX := startPoint.X
// 	sY := startPoint.Y
// 	c1X := float32(control1.X)
// 	c1Y := float32(control1.Y)
// 	c2X := float32(control2.X)
// 	c2Y := float32(control2.Y)
// 	eX := float32(position.X)
// 	eY := float32(position.Y)

// 	const precision = 30 // TODO: Evaluate based on points
// 	for i := 1; i <= precision; i++ {
// 		// TODO: Use derivatives for performance improvement
// 		t := float32(i) / float32(precision)
// 		alpha := (1 - t) * (1 - t) * (1 - t)
// 		beta := 3 * (1 - t) * (1 - t) * t
// 		gamma := 3 * (1 - t) * t * t
// 		delta := t * t * t
// 		c.activeShape.Points = append(c.activeShape.Points, Point{
// 			Vec2: sprec.NewVec2(
// 				alpha*sX+beta*c1X+gamma*c2X+delta*eX,
// 				alpha*sY+beta*c1Y+gamma*c2Y+delta*eY,
// 			),
// 			InStroke:  endStroke, // TODO: Interpolate
// 			OutStroke: endStroke, // TODO: Interpolate and set only if non-last
// 		})
// 	}
// }

// func (c *Canvas) CloseLoop(startStroke, endStroke ui.Stroke) {
// 	c.activeShape.Points[len(c.activeShape.Points)-1].OutStroke = startStroke
// 	c.activeShape.Points = append(c.activeShape.Points, Point{
// 		Vec2:     c.activeShape.Points[0].Vec2,
// 		InStroke: endStroke,
// 	})
// }

// func (c *Canvas) EndShape() {
// 	translation := sprec.NewVec2(
// 		float32(c.currentLayer.Translation.X),
// 		float32(c.currentLayer.Translation.Y),
// 	)

// 	offset := c.mesh.Offset()
// 	// TODO: Only if background color or texture is set
// 	for _, point := range c.activeShape.Points {
// 		c.mesh.Append(Vertex{
// 			position: sprec.Vec2Sum(point.Vec2, translation),
// 			texCoord: sprec.NewVec2(0.0, 0.0),
// 			color:    c.activeShape.BackgroundColor,
// 		})
// 	}
// 	count := c.mesh.Offset() - offset

// 	cullFace := gl.BACK
// 	if c.activeShape.Winding == ui.WindingCW {
// 		cullFace = gl.FRONT
// 	}

// 	if c.activeShape.Rule == ui.FillRuleSimple {
// 		c.subMeshes = append(c.subMeshes, SubMesh{
// 			clipBounds: sprec.NewVec4(
// 				float32(c.currentLayer.ClipBounds.X),
// 				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
// 				float32(c.currentLayer.ClipBounds.Y),
// 				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
// 			),
// 			material:     c.opaqueMaterial,
// 			texture:      c.whiteMask,
// 			vertexOffset: offset,
// 			vertexCount:  count,
// 			cullFace:     uint32(cullFace),
// 			primitive:    gl.TRIANGLE_FAN,
// 		})
// 	}

// 	if c.activeShape.Rule != ui.FillRuleSimple {
// 		// clear stencil
// 		c.subMeshes = append(c.subMeshes, SubMesh{
// 			clipBounds: sprec.NewVec4(
// 				float32(c.currentLayer.ClipBounds.X),
// 				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
// 				float32(c.currentLayer.ClipBounds.Y),
// 				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
// 			),
// 			material:     c.opaqueMaterial,
// 			texture:      c.whiteMask,
// 			vertexOffset: offset,
// 			vertexCount:  count,
// 			culling:      false,
// 			cullFace:     uint32(cullFace),
// 			primitive:    gl.TRIANGLE_FAN,
// 			skipColor:    true,
// 			stencil:      true,
// 			stencilCfg: stencilConfig{
// 				stencilFuncFront: stencilFunc{
// 					fn:   gl.ALWAYS,
// 					ref:  0,
// 					mask: 0xFF,
// 				},
// 				stencilFuncBack: stencilFunc{
// 					fn:   gl.ALWAYS,
// 					ref:  0,
// 					mask: 0xFF,
// 				},
// 				stencilOpFront: stencilOp{
// 					sfail:  gl.REPLACE,
// 					dpfail: gl.REPLACE,
// 					dppass: gl.REPLACE,
// 				},
// 				stencilOpBack: stencilOp{
// 					sfail:  gl.REPLACE,
// 					dpfail: gl.REPLACE,
// 					dppass: gl.REPLACE,
// 				},
// 			},
// 		})

// 		// render stencil mask
// 		c.subMeshes = append(c.subMeshes, SubMesh{
// 			clipBounds: sprec.NewVec4(
// 				float32(c.currentLayer.ClipBounds.X),
// 				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
// 				float32(c.currentLayer.ClipBounds.Y),
// 				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
// 			),
// 			material:     c.opaqueMaterial,
// 			texture:      c.whiteMask,
// 			vertexOffset: offset,
// 			vertexCount:  count,
// 			cullFace:     uint32(cullFace),
// 			primitive:    gl.TRIANGLE_FAN,
// 			skipColor:    true, // we don't want to render anything
// 			stencil:      true,
// 			stencilCfg: stencilConfig{
// 				stencilFuncFront: stencilFunc{
// 					fn:   gl.ALWAYS,
// 					ref:  0,
// 					mask: 0xFF,
// 				},
// 				stencilFuncBack: stencilFunc{
// 					fn:   gl.ALWAYS,
// 					ref:  0,
// 					mask: 0xFF,
// 				},
// 				stencilOpFront: stencilOp{
// 					sfail:  gl.KEEP,
// 					dpfail: gl.KEEP,
// 					dppass: gl.INCR_WRAP, // increase correct winding
// 				},
// 				stencilOpBack: stencilOp{
// 					sfail:  gl.KEEP,
// 					dpfail: gl.KEEP,
// 					dppass: gl.DECR_WRAP, // decrease incorrect winding
// 				},
// 			},
// 		})

// 		// render final polygon
// 		c.subMeshes = append(c.subMeshes, SubMesh{
// 			clipBounds: sprec.NewVec4(
// 				float32(c.currentLayer.ClipBounds.X),
// 				float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
// 				float32(c.currentLayer.ClipBounds.Y),
// 				float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
// 			),
// 			material:     c.opaqueMaterial,
// 			texture:      c.whiteMask,
// 			vertexOffset: offset,
// 			vertexCount:  count,
// 			cullFace:     uint32(cullFace),
// 			primitive:    gl.TRIANGLE_FAN,
// 			skipColor:    false, // we want to render now
// 			stencil:      true,
// 			stencilCfg: stencilConfig{
// 				stencilFuncFront: stencilFunc{
// 					fn:   gl.LESS,
// 					ref:  0,
// 					mask: 0xFF,
// 				},
// 				stencilFuncBack: stencilFunc{
// 					fn:   gl.LESS,
// 					ref:  0,
// 					mask: 0xFF,
// 				},
// 				stencilOpFront: stencilOp{
// 					sfail:  gl.KEEP,
// 					dpfail: gl.KEEP,
// 					dppass: gl.KEEP,
// 				},
// 				stencilOpBack: stencilOp{
// 					sfail:  gl.KEEP,
// 					dpfail: gl.KEEP,
// 					dppass: gl.KEEP,
// 				},
// 			},
// 		})
// 	}

// 	// TODO: Draw stroke
// }
