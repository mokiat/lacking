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
		canvasPath: newCanvasPath(),
		api:        api,
		state:      state,
		shape:      newShape(state, shaders),
		contour:    newContour(state, shaders),
		text:       newText(state, shaders),
	}
}

type canvasRenderer struct {
	*canvasPath
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

// Push records the current state and creates a new state layer. Changes done
// in the new layer will not affect the parent layer.
//
// You may create up to 256 layers including the starting one after which the
// method panics.
func (c *canvasRenderer) Push() {
	c.state.currentLayer = c.state.currentLayer.Next()
}

// Pop restores the drawing state based on the parent layer. If this is the
// first layer, then this method panics.
func (c *canvasRenderer) Pop() {
	c.state.currentLayer = c.state.currentLayer.Previous()
}

// ResetTransform restores the transform to the value it had
// after the last Push. If this is the first layer, then it is
// set to the identity matrix.
func (c *canvasRenderer) ResetTransform() {
	if c.state.currentLayer == c.state.topLayer {
		c.state.currentLayer.Transform = sprec.IdentityMat4()
	} else {
		c.state.currentLayer.Transform = c.state.currentLayer.previous.Transform
	}
}

// SetTransform changes the transform relative to the former layer transform.
func (c *canvasRenderer) SetTransform(transform sprec.Mat4) {
	if c.state.currentLayer == c.state.topLayer {
		c.state.currentLayer.Transform = transform
	} else {
		c.state.currentLayer.Transform = sprec.Mat4Prod(
			c.state.currentLayer.previous.Transform,
			transform,
		)
	}
}

// Translate moves the drawing position by the specified delta amount.
func (c *canvasRenderer) Translate(delta sprec.Vec2) {
	c.state.currentLayer.Transform = sprec.Mat4Prod(
		c.state.currentLayer.Transform,
		sprec.TranslationMat4(delta.X, delta.Y, 0.0),
	)
}

func (c *canvasRenderer) SetClipBounds(left, right, top, bottom float32) {
	c.state.currentLayer.ClipBounds = NewBounds(
		int(left),
		int(top),
		int(right-left),
		int(bottom-top),
	)
}

// Clip sets new clipping bounds. Pixels from draw operations
// that are outside the clipping bounds will not be drawn.
//
// Initially the clipping bounds are equal to the window size.
func (c *canvasRenderer) Clip(bounds Bounds) {
	// FIXME: This no longer works correctly
	c.state.currentLayer.ClipBounds = bounds.Translate(
		NewPosition(
			int(c.state.currentLayer.Transform.Translation().X),
			int(c.state.currentLayer.Transform.Translation().Y),
		),
	)
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
func (c *canvasRenderer) DrawSurface(surface Surface, position Position, size Size) {
	texture := surface.Render(size.Width, size.Height)

	c.Reset()
	c.Rectangle(
		sprec.NewVec2(float32(position.X), float32(position.Y)),
		sprec.NewVec2(float32(size.Width), float32(size.Height)),
	)
	c.Fill(Fill{
		Rule: FillRuleSimple,
		Image: &Image{ // TODO: Don't allocate
			texture: texture,
			size:    size,
		},
		Color:       White(),
		ImageOffset: sprec.NewVec2(0.0, float32(size.Height)),
		ImageSize:   sprec.NewVec2(float32(size.Width), -float32(size.Height)),
	})
}

func (c *canvasRenderer) Fill(fill Fill) {
	c.fillPath(c.canvasPath, fill)
}

func (c *canvasRenderer) Stroke() {
	c.strokePath(c.canvasPath)
}

func (c *canvasRenderer) fillPath(path *canvasPath, fill Fill) {
	// TODO: Implement directly and remove old API
	c.Shape().Begin(fill)
	for i := 0; i < len(path.subPathOffsets); i++ {
		offset := path.subPathOffsets[i]
		nextOffset := len(path.points)
		if i+1 < len(path.subPathOffsets) {
			nextOffset = path.subPathOffsets[i+1]
		}
		for j, point := range path.points[offset:nextOffset] {
			if j == 0 {
				c.Shape().MoveTo(point.coords)
			} else {
				c.Shape().LineTo(point.coords)
			}
		}
	}
	c.Shape().End()
}

func (c *canvasRenderer) strokePath(path *canvasPath) {
	// TODO: Implement directly and remove old API
	c.Contour().Begin()
	for i := 0; i < len(path.subPathOffsets); i++ {
		offset := path.subPathOffsets[i]
		nextOffset := len(path.points)
		if i+1 < len(path.subPathOffsets) {
			nextOffset = path.subPathOffsets[i+1]
		}
		for j, point := range path.points[offset:nextOffset] {
			if j == 0 {
				c.Contour().MoveTo(point.coords, Stroke{
					Size:  point.innerSize + point.outerSize,
					Color: point.color,
				})
			} else {
				c.Contour().LineTo(point.coords, Stroke{
					Size:  point.innerSize + point.outerSize,
					Color: point.color,
				})
			}
		}
	}
	c.Contour().End()
}

type Surface interface {
	Render(width, height int) render.Texture
}
