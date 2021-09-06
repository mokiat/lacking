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

func (c *Canvas) DrawText(value string, position ui.Position) {
	c.renderer.SetClipBounds(
		float32(c.currentLayer.ClipBounds.X),
		float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
		float32(c.currentLayer.ClipBounds.Y),
		float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
	)
	c.renderer.SetTransform(sprec.TranslationMat4(
		float32(c.currentLayer.Translation.X+position.X),
		float32(c.currentLayer.Translation.Y+position.Y),
		0.0,
	))

	font := c.currentLayer.Font
	fontSize := c.currentLayer.FontSize
	color := sprec.NewVec4(
		float32(c.currentLayer.SolidColor.R)/255.0,
		float32(c.currentLayer.SolidColor.G)/255.0,
		float32(c.currentLayer.SolidColor.B)/255.0,
		float32(c.currentLayer.SolidColor.A)/255.0,
	)

	text := c.renderer.BeginText(font, fontSize, color)
	text.Write(value)
	c.renderer.EndText(text)
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
