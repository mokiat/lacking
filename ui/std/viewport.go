package std

import (
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

var (
	ViewportInitialFBWidth  = 400
	ViewportInitialFBHeight = 300
)

// Viewport represents a component that uses low-level API calls to render
// an external scene, unlike other components that rely on the Canvas API.
var Viewport = co.Define(&viewportComponent{})

// ViewportData holds the data for a Viewport component.
type ViewportData struct {
	API render.API
}

// ViewportCallbackData holds the callback data for a Viewport component.
type ViewportCallbackData struct {
	OnKeyboardEvent func(element *ui.Element, event ui.KeyboardEvent) bool
	OnMouseEvent    func(element *ui.Element, event ui.MouseEvent) bool
	OnRender        func(framebuffer render.Framebuffer, fbSize ui.Size)
}

type viewportComponent struct {
	co.BaseComponent

	surface viewportSurface

	fbResizeBlocked bool
	fbDesiredWidth  int
	fbDesiredHeight int

	onKeyboardEvent func(element *ui.Element, event ui.KeyboardEvent) bool
	onMouseEvent    func(element *ui.Element, event ui.MouseEvent) bool
	onRender        func(framebuffer render.Framebuffer, fbSize ui.Size)
}

func (c *viewportComponent) OnCreate() {
	data := co.GetData[ViewportData](c.Properties())
	c.surface.api = data.API

	callbackData := co.GetOptionalCallbackData(c.Properties(), ViewportCallbackData{})
	c.onKeyboardEvent = callbackData.OnKeyboardEvent
	if c.onKeyboardEvent == nil {
		c.onKeyboardEvent = func(element *ui.Element, event ui.KeyboardEvent) bool {
			return false
		}
	}
	c.onMouseEvent = callbackData.OnMouseEvent
	if c.onMouseEvent == nil {
		c.onMouseEvent = func(element *ui.Element, event ui.MouseEvent) bool {
			return false
		}
	}
	c.onRender = callbackData.OnRender
	if c.onRender == nil {
		c.onRender = func(framebuffer render.Framebuffer, fbSize ui.Size) {}
	}

	c.surface.onRender = c.onRender
	c.surface.createFramebuffer(ViewportInitialFBWidth, ViewportInitialFBHeight)
}

func (c *viewportComponent) OnDelete() {
	c.surface.releaseFramebuffer()
}

func (c *viewportComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence:   c,
			Focusable: opt.V(true),
			Focused:   opt.V(true),
		})
		co.WithChildren(c.Properties().Children())
	})
}

func (c *viewportComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	return c.onKeyboardEvent(element, event)
}

func (c *viewportComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	return c.onMouseEvent(element, event)
}

func (c *viewportComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	size := element.Bounds().Size
	c.fbDesiredWidth = size.Width
	c.fbDesiredHeight = size.Height

	if c.fbDesiredWidth != c.surface.fbWidth || c.fbDesiredHeight != c.surface.fbHeight {
		if !c.fbResizeBlocked {
			c.surface.releaseFramebuffer()
			c.surface.createFramebuffer(c.fbDesiredWidth, c.fbDesiredHeight)
			c.fbResizeBlocked = true

			co.After(c.Scope(), time.Second, func() {
				c.fbResizeBlocked = false
				element.Invalidate()
			})
		}
	}

	drawBounds := canvas.DrawBounds(element, false)

	if c.surface.onRender != nil {
		canvas.DrawSurface(
			&c.surface,
			drawBounds.Position,
			drawBounds.Size,
		)
	} else {
		canvas.Reset()
		canvas.Rectangle(
			drawBounds.Position,
			drawBounds.Size,
		)
		canvas.Fill(ui.Fill{
			Rule:  ui.FillRuleSimple,
			Color: BackgroundColor,
		})
	}
}

type viewportSurface struct {
	api          render.API
	colorTexture render.Texture
	framebuffer  render.Framebuffer

	fbWidth  int
	fbHeight int

	onRender func(framebuffer render.Framebuffer, size ui.Size)
}

func (c *viewportSurface) Render() (render.Texture, ui.Size) {
	fbSize := ui.NewSize(c.fbWidth, c.fbHeight)
	c.onRender(c.framebuffer, fbSize)
	return c.colorTexture, fbSize
}

func (c *viewportSurface) createFramebuffer(width, height int) {
	c.fbWidth = width
	c.fbHeight = height
	c.colorTexture = c.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           width,
		Height:          height,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
	})
	c.framebuffer = c.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			c.colorTexture,
		},
	})
}

func (c *viewportSurface) releaseFramebuffer() {
	if c.framebuffer != nil {
		c.framebuffer.Release()
		c.framebuffer = nil
	}
	if c.colorTexture != nil {
		c.colorTexture.Release()
		c.colorTexture = nil
	}
}
