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

// ViewportMouseEvent is a wrapper on top of ui.MouseEvent which includes
// X and Y coordinates that are in the [0..1] range, making them size agnostic.
type ViewportMouseEvent struct {
	ui.MouseEvent
	X float32
	Y float32
}

// ViewportData holds the data for a Viewport component.
type ViewportData struct {
	API render.API
}

// ViewportCallbackData holds the callback data for a Viewport component.
type ViewportCallbackData struct {
	OnKeyboardEvent func(event ui.KeyboardEvent) bool
	OnMouseEvent    func(event ViewportMouseEvent) bool
	OnRender        func(framebuffer render.Framebuffer, size ui.Size)
}

var viewportDefaultCallbackData = ViewportCallbackData{
	OnKeyboardEvent: func(event ui.KeyboardEvent) bool {
		return false
	},
	OnMouseEvent: func(event ViewportMouseEvent) bool {
		return false
	},
	OnRender: func(framebuffer render.Framebuffer, size ui.Size) {},
}

// Viewport represents a component that uses low-level API calls to render
// an external scene, unlike other components that rely on the Canvas API.
var Viewport = co.Define(&viewportComponent{})

type viewportComponent struct {
	co.BaseComponent

	surface viewportSurfaceComponent

	fbResizeBlocked bool
	fbDesiredWidth  int
	fbDesiredHeight int

	onKeyboardEvent func(event ui.KeyboardEvent) bool
	onMouseEvent    func(event ViewportMouseEvent) bool
}

func (c *viewportComponent) OnCreate() {
	data := co.GetData[ViewportData](c.Properties())
	c.surface.api = data.API

	c.surface.createFramebuffer(ViewportInitialFBWidth, ViewportInitialFBHeight)
}

func (c *viewportComponent) OnUpsert() {
	data := co.GetData[ViewportData](c.Properties())
	c.surface.api = data.API

	callbackData := co.GetOptionalCallbackData(c.Properties(), viewportDefaultCallbackData)
	c.onKeyboardEvent = callbackData.OnKeyboardEvent
	c.onMouseEvent = callbackData.OnMouseEvent
	c.surface.onRender = callbackData.OnRender
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
	return c.onKeyboardEvent(event)
}

func (c *viewportComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	size := element.Bounds().Size
	x := -1.0 + 2.0*float32(event.Position.X)/float32(size.Width)
	y := 1.0 - 2.0*float32(event.Position.Y)/float32(size.Height)
	return c.onMouseEvent(ViewportMouseEvent{
		MouseEvent: event,
		X:          x,
		Y:          y,
	})
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

			time.AfterFunc(time.Second, func() {
				element.Window().Schedule(func() error {
					c.fbResizeBlocked = false
					element.Invalidate()
					return nil
				})
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

type viewportSurfaceComponent struct {
	api          render.API
	colorTexture render.Texture
	framebuffer  render.Framebuffer

	fbWidth  int
	fbHeight int

	onRender func(framebuffer render.Framebuffer, size ui.Size)
}

func (c *viewportSurfaceComponent) Render() (render.Texture, ui.Size) {
	fbSize := ui.NewSize(c.fbWidth, c.fbHeight)
	c.onRender(c.framebuffer, fbSize)
	return c.colorTexture, fbSize
}

func (c *viewportSurfaceComponent) createFramebuffer(width, height int) {
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

func (c *viewportSurfaceComponent) releaseFramebuffer() {
	if c.framebuffer != nil {
		c.framebuffer.Release()
		c.framebuffer = nil
	}
	if c.colorTexture != nil {
		c.colorTexture.Release()
		c.colorTexture = nil
	}
}
