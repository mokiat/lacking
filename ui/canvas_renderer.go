package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

const (
	maxLayerDepth = 256

	maxVertexCount = 524288
)

func newCanvasRenderer(api render.API, shaders ShaderCollection) *canvasRenderer {
	state := newCanvasState()
	return &canvasRenderer{
		api:     api,
		state:   state,
		shape:   newShape(state, shaders),
		contour: newContour(state, shaders),
		text:    newText(state, shaders),
	}
}

type canvasRenderer struct {
	api          render.API
	state        *canvasState
	commandQueue render.CommandQueue

	shape   *Shape
	contour *Contour
	text    *Text
}

func (c *canvasRenderer) onCreate() {
	c.commandQueue = c.api.CreateCommandQueue()
	c.state.onCreate(c.api)
	c.shape.onCreate(c.api, c.commandQueue)
	c.contour.onCreate(c.api, c.commandQueue)
	c.text.onCreate(c.api, c.commandQueue)
}

func (c *canvasRenderer) onDestroy() {
	defer c.commandQueue.Release()
	defer c.state.onDestroy()
	defer c.shape.onDestroy()
	defer c.contour.onDestroy()
	defer c.text.onDestroy()
}

func (c *canvasRenderer) onBegin(size Size) {
	c.state.currentLayer = c.state.topLayer
	c.state.currentLayer.ClipBounds.Position = NewPosition(0, 0)
	c.state.currentLayer.ClipBounds.Size = size
	c.state.currentLayer.Transform = sprec.IdentityMat4()
	c.state.projectionMatrix = sprec.OrthoMat4(
		0.0, float32(size.Width),
		0.0, float32(size.Height),
		0.0, 1.0,
	)
	c.shape.onBegin()
	c.contour.onBegin()
	c.text.onBegin()
}

func (c *canvasRenderer) onEnd() {
	c.shape.onEnd()
	c.contour.onEnd()
	c.text.onEnd()
	c.api.SubmitQueue(c.commandQueue)
}

// Push records the current state and creates a new
// state layer. Changes done in the new layer will
// not affect the former layer.
func (c *canvasRenderer) Push() {
	c.state.currentLayer = c.state.currentLayer.Next()
}

// Pop restores the former state layer and configures
// the drawing state accordingly.
func (c *canvasRenderer) Pop() {
	c.state.currentLayer = c.state.currentLayer.Previous()
}

func (c *canvasRenderer) SetClipBounds(left, right, top, bottom float32) {
	c.state.currentLayer.ClipBounds = NewBounds(
		int(left),
		int(top),
		int(right-left),
		int(bottom-top),
	)
}

func (c *canvasRenderer) SetTransform(transform sprec.Mat4) {
	c.state.currentLayer.Transform = transform
}

// Translate moves the drawing position by the specified
// delta amount.
func (c *canvasRenderer) Translate(delta Position) {
	// FIXME
	c.state.currentLayer.Transform = sprec.Mat4MultiProd(
		sprec.TranslationMat4(float32(delta.X), float32(delta.Y), 0.0),
		c.state.currentLayer.Transform,
	)
	// c.state.currentLayer.Translation = c.state.currentLayer.Translation.Translate(delta.X, delta.Y)
}

// Clip sets new clipping bounds. Pixels from draw operations
// that are outside the clipping bounds will not be drawn.
//
// Initially the clipping bounds are equal to the window size.
func (c *canvasRenderer) Clip(bounds Bounds) {
	// FIXME
	// if previousLayer := c.state.currentLayer.previous; previousLayer != nil {
	// 		previousClipBounds := previousLayer.ClipBounds
	// 		newClipBounds := bounds.Translate(c.state.currentLayer.Translation)
	// 		c.state.currentLayer.ClipBounds = previousClipBounds.Intersect(newClipBounds)
	// } else {
	c.state.currentLayer.ClipBounds = bounds.Translate(
		NewPosition(
			int(c.state.currentLayer.Transform.Translation().X),
			int(c.state.currentLayer.Transform.Translation().Y),
		),
	)
	// }
}

// Shape returns the shape rendering module.
func (c *canvasRenderer) Shape() *Shape {
	return c.shape
}

// Contour returns the contour rendering module.
func (c *canvasRenderer) Contour() *Contour {
	return c.contour
}

// Text returns the text rendering module.
func (c *canvasRenderer) Text() *Text {
	return c.text
}

// DrawSurface renders the specified surface. The surface's Render
// method will be called when needed with the UI framebuffer bound.
func (c *canvasRenderer) DrawSurface(surface Surface) {
	// TODO
	// currentLayer := c.currentLayer
	// c.renderer.SetClipBounds(
	// 	float32(currentLayer.ClipBounds.X),
	// 	float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
	// 	float32(currentLayer.ClipBounds.Y),
	// 	float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	// )
	// c.renderer.SetTransform(sprec.TranslationMat4(
	// 	float32(currentLayer.Translation.X),
	// 	float32(currentLayer.Translation.Y),
	// 	0.0,
	// ))
	// c.renderer.DrawSurface(surface)
}

// RectRoundness is used to configure the roundness of
// a round rectangle through corner radiuses.
type RectRoundness struct {

	// TopLeftRadius specifies the radius of the top-left corner.
	TopLeftRadius float32

	// TopRightRadius specifies the radius of the top-right corner.
	TopRightRadius float32

	// BottomLeftRadius specifies the radius of the bottom-left corner.
	BottomLeftRadius float32

	// BottomRightRadius specifies the radius of the bottom-right corner.
	BottomRightRadius float32
}

type Surface interface {
	Render(x, y, width, height int)
}
