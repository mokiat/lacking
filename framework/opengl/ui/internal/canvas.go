package internal

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

func NewCanvas(renderer *Renderer) *Canvas {
	result := &Canvas{
		renderer: renderer,

		defaultLayer: &Layer{
			Translation: ui.NewPosition(0, 0),
			ClipBounds:  ui.NewBounds(0, 0, 1, 1),
		},
		topLayer: &Layer{},

		framebuffer: opengl.DefaultFramebuffer(),
	}
	result.shape = &canvasShape{
		canvas:   result,
		renderer: renderer,
	}
	result.contour = &canvasContour{
		canvas:   result,
		renderer: renderer,
	}
	result.text = &canvasText{
		canvas:   result,
		renderer: renderer,
	}
	return result
}

var _ ui.Canvas = (*Canvas)(nil)

type Canvas struct {
	renderer *Renderer

	defaultLayer *Layer
	topLayer     *Layer
	currentLayer *Layer

	framebuffer *opengl.Framebuffer
	windowSize  ui.Size

	shape   *canvasShape
	contour *canvasContour
	text    *canvasText
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

	gl.Enable(gl.FRAMEBUFFER_SRGB)
	c.framebuffer.ClearDepth(1.0)

	c.renderer.Begin(Target{
		Framebuffer: c.framebuffer,
		Width:       c.windowSize.Width,
		Height:      c.windowSize.Height,
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
	if previousLayer := c.currentLayer.previous; previousLayer != nil {
		previousClipBounds := previousLayer.ClipBounds
		newClipBounds := bounds.Translate(c.currentLayer.Translation)
		c.currentLayer.ClipBounds = previousClipBounds.Intersect(newClipBounds)
	} else {
		c.currentLayer.ClipBounds = bounds.Translate(c.currentLayer.Translation)
	}
}

func (c *Canvas) Shape() ui.Shape {
	return c.shape
}

func (c *Canvas) Contour() ui.Contour {
	return c.contour
}

func (c *Canvas) Text() ui.Text {
	return c.text
}

func (c *Canvas) DrawSurface(surface ui.Surface) {
	currentLayer := c.currentLayer
	c.renderer.SetClipBounds(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	c.renderer.SetTransform(sprec.TranslationMat4(
		float32(currentLayer.Translation.X),
		float32(currentLayer.Translation.Y),
		0.0,
	))
	c.renderer.DrawSurface(surface)
}

var _ ui.Shape = (*canvasShape)(nil)

type canvasShape struct {
	renderer *Renderer
	canvas   *Canvas
	shape    *Shape
}

func (s *canvasShape) Begin(fill ui.Fill) {
	currentLayer := s.canvas.currentLayer
	s.renderer.SetClipBounds(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	s.renderer.SetTransform(sprec.TranslationMat4(
		float32(currentLayer.Translation.X),
		float32(currentLayer.Translation.Y),
		0.0,
	))
	s.renderer.SetTextureTransform(sprec.Mat4MultiProd(
		sprec.ScaleMat4(
			1.0/float32(fill.ImageSize.Width),
			1.0/float32(fill.ImageSize.Height),
			1.0,
		),
		sprec.TranslationMat4(
			-float32(fill.ImageOffset.X),
			-float32(fill.ImageOffset.Y),
			0.0,
		),
	))
	s.shape = s.renderer.BeginShape(Fill{
		mode:  uiFillRuleToStencilMode(fill.Rule),
		color: uiColorToVec(fill.Color),
		image: uiImageToImage(fill.Image),
	})
}

func (s *canvasShape) MoveTo(position ui.Position) {
	s.shape.MoveTo(sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	))
}

func (s *canvasShape) LineTo(position ui.Position) {
	s.shape.LineTo(sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	))
}

func (s *canvasShape) QuadTo(control, position ui.Position) {
	s.shape.QuadTo(sprec.NewVec2(
		float32(control.X),
		float32(control.Y),
	), sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	))
}

func (s *canvasShape) CubeTo(control1, control2, position ui.Position) {
	s.shape.CubeTo(sprec.NewVec2(
		float32(control1.X),
		float32(control1.Y),
	), sprec.NewVec2(
		float32(control2.X),
		float32(control2.Y),
	), sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	))
}

func (s *canvasShape) Rectangle(position ui.Position, size ui.Size) {
	s.MoveTo(position)
	s.LineTo(ui.NewPosition(
		position.X,
		position.Y+size.Height,
	))
	s.LineTo(ui.NewPosition(
		position.X+size.Width,
		position.Y+size.Height,
	))
	s.LineTo(ui.NewPosition(
		position.X+size.Width,
		position.Y,
	))
}

func (s *canvasShape) Triangle(a, b, c ui.Position) {
	s.MoveTo(a)
	s.LineTo(b)
	s.LineTo(c)
}

func (s *canvasShape) Circle(position ui.Position, radius int) {
	// TODO
}

func (s *canvasShape) RoundRectangle(position ui.Position, size ui.Size, roundness ui.RectRoundness) {
	// TODO
}

func (s *canvasShape) End() {
	s.renderer.EndShape(s.shape)
}

var _ ui.Contour = (*canvasContour)(nil)

type canvasContour struct {
	canvas   *Canvas
	renderer *Renderer
	contour  *Contour
}

func (c *canvasContour) Begin() {
	currentLayer := c.canvas.currentLayer
	c.renderer.SetClipBounds(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	c.renderer.SetTransform(sprec.TranslationMat4(
		float32(currentLayer.Translation.X),
		float32(currentLayer.Translation.Y),
		0.0,
	))
	c.contour = c.renderer.BeginContour()
}

func (c *canvasContour) MoveTo(position ui.Position, stroke ui.Stroke) {
	c.contour.MoveTo(
		uiPositionToVec(position),
		uiStrokeToStroke(stroke),
	)
}

func (c *canvasContour) LineTo(position ui.Position, stroke ui.Stroke) {
	c.contour.LineTo(
		uiPositionToVec(position),
		uiStrokeToStroke(stroke),
	)
}

func (c *canvasContour) QuadTo(control, position ui.Position, stroke ui.Stroke) {
	c.contour.QuadTo(
		uiPositionToVec(control),
		uiPositionToVec(position),
		uiStrokeToStroke(stroke),
	)
}

func (c *canvasContour) CubeTo(control1, control2, position ui.Position, stroke ui.Stroke) {
	c.contour.CubeTo(
		uiPositionToVec(control1),
		uiPositionToVec(control2),
		uiPositionToVec(position),
		uiStrokeToStroke(stroke),
	)
}

func (c *canvasContour) CloseLoop() {
	c.contour.CloseLoop()
}

func (c *canvasContour) Rectangle(position ui.Position, size ui.Size, stroke ui.Stroke) {
	c.MoveTo(position, stroke)
	c.LineTo(ui.NewPosition(
		position.X,
		position.Y+size.Height,
	), stroke)
	c.LineTo(ui.NewPosition(
		position.X+size.Width,
		position.Y+size.Height,
	), stroke)
	c.LineTo(ui.NewPosition(
		position.X+size.Width,
		position.Y,
	), stroke)
	c.CloseLoop()
}

func (c *canvasContour) Triangle(p1, p2, p3 ui.Position, stroke ui.Stroke) {
	c.MoveTo(p1, stroke)
	c.LineTo(p2, stroke)
	c.LineTo(p3, stroke)
	c.CloseLoop()
}

func (c *canvasContour) Circle(position ui.Position, radius int, stroke ui.Stroke) {
	// TODO
}

func (c *canvasContour) RoundRectangle(position ui.Position, size ui.Size, roundness ui.RectRoundness, stroke ui.Stroke) {
	// TODO
}

func (c *canvasContour) End() {
	c.renderer.EndContour(c.contour)
}

var _ ui.Text = (*canvasText)(nil)

type canvasText struct {
	canvas   *Canvas
	renderer *Renderer
	text     *Text
}

func (t *canvasText) Begin(typography ui.Typography) {
	currentLayer := t.canvas.currentLayer
	t.renderer.SetClipBounds(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	t.renderer.SetTransform(sprec.TranslationMat4(
		float32(currentLayer.Translation.X),
		float32(currentLayer.Translation.Y),
		0.0,
	))
	t.text = t.renderer.BeginText(Typography{
		Font:  uiFontToFont(typography.Font),
		Size:  float32(typography.Size),
		Color: uiColorToVec(typography.Color),
	})
}

func (t *canvasText) Line(value string, position ui.Position) {
	t.text.Write(value, sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	))
}

func (t *canvasText) End() {
	t.renderer.EndText(t.text)
}

func uiPositionToVec(position ui.Position) sprec.Vec2 {
	return sprec.NewVec2(
		float32(position.X),
		float32(position.Y),
	)
}

func uiColorToVec(color ui.Color) sprec.Vec4 {
	return sprec.NewVec4(
		float32(color.R)/255.0,
		float32(color.G)/255.0,
		float32(color.B)/255.0,
		float32(color.A)/255.0,
	)
}

func uiImageToImage(image ui.Image) *Image {
	if image == nil {
		return nil
	}
	return image.(*Image)
}

func uiFillRuleToStencilMode(rule ui.FillRule) StencilMode {
	switch rule {
	case ui.FillRuleSimple:
		return StencilModeNone
	case ui.FillRuleNonZero:
		return StencilModeNonZero
	case ui.FillRuleEvenOdd:
		return StencilModeOdd
	default:
		return StencilModeNone
	}
}

func uiStrokeToStroke(stroke ui.Stroke) Stroke {
	return Stroke{
		innerSize: float32(stroke.Size),
		color:     uiColorToVec(stroke.Color),
	}
}

func uiFontToFont(font ui.Font) *Font {
	if font == nil {
		return nil
	}
	return font.(*Font)
}
