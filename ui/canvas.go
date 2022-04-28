package ui

import (
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
}

func (c *Canvas) onResize(size Size) {
	c.windowSize = size
}

func (c *Canvas) onResizeFramebuffer(size Size) {
	// TODO: Use own framebuffer which would allow for
	// only dirty region rerendering even when overlay.
}

func (c *Canvas) onBegin() {
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
