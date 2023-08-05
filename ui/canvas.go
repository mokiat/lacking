package ui

import (
	"time"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

func newCanvas(renderer *canvasRenderer) *Canvas {
	return &Canvas{
		canvasRenderer: renderer,
		framebuffer:    renderer.api.DefaultFramebuffer(),
	}
}

// Canvas represents a mechanism through which an Element can render itself
// to the screen.
type Canvas struct {
	*canvasRenderer

	framebuffer render.Framebuffer
	windowSize  Size
	deltaTime   time.Duration
}

// ElapsedTime returns the amount of time that has passed since the last
// render iteration.
//
// This should only be used by elements that are constantly being invalidated
// (i.e. do real-time rendering), as otherwise this duration would be
// incorrect since a non-dirty element could be omitted during some frames.
func (c *Canvas) ElapsedTime() time.Duration {
	return c.deltaTime
}

// DrawBounds returns the bounds to be used for drawing for the specified
// element.
func (c *Canvas) DrawBounds(element *Element, padding bool) DrawBounds {
	if !padding {
		size := element.Bounds().Size
		return DrawBounds{
			Position: sprec.ZeroVec2(),
			Size:     sprec.NewVec2(float32(size.Width), float32(size.Height)),
		}
	}
	contentBounds := element.ContentBounds()
	return DrawBounds{
		Position: sprec.NewVec2(float32(contentBounds.X), float32(contentBounds.Y)),
		Size:     sprec.NewVec2(float32(contentBounds.Width), float32(contentBounds.Height)),
	}
}

func (c *Canvas) onResize(size Size) {
	c.windowSize = size
}

func (c *Canvas) onResizeFramebuffer(size Size) {
	// TODO: Use own framebuffer which would allow for
	// only dirty region rerendering even when overlay.
}

func (c *Canvas) onBegin(deltaTime time.Duration) {
	c.deltaTime = deltaTime
	c.canvasRenderer.onBegin(c.windowSize)
}

func (c *Canvas) onEnd() {
	c.api.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: c.framebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  c.windowSize.Width,
			Height: c.windowSize.Height,
		},
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:  render.LoadOperationDontCare,
				StoreOp: render.StoreOperationStore,
			},
		},
		DepthLoadOp:       render.LoadOperationDontCare,
		DepthStoreOp:      render.StoreOperationDontCare,
		StencilLoadOp:     render.LoadOperationClear,
		StencilStoreOp:    render.StoreOperationDontCare,
		StencilClearValue: 0x00,
	})
	c.canvasRenderer.onEnd()
	c.api.EndRenderPass()
}

// DrawBounds represents a rectangle area to be used for drawing.
type DrawBounds struct {
	Position sprec.Vec2
	Size     sprec.Vec2
}

// X returns the left side of the draw area.
func (b DrawBounds) X() float32 {
	return b.Position.X
}

// Y returns the top side of the draw area.
func (b DrawBounds) Y() float32 {
	return b.Position.Y
}

// Width returns the width of the draw area.
func (b DrawBounds) Width() float32 {
	return b.Size.X
}

// Height returns the height of the draw area.
func (b DrawBounds) Height() float32 {
	return b.Size.Y
}
