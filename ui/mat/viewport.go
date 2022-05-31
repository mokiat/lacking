package mat

import (
	"time"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

var (
	ViewportInitialFBWidth  = 400
	ViewportInitialFBHeight = 300
)

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

var defaultViewportCallbackData = ViewportCallbackData{
	OnKeyboardEvent: func(event ui.KeyboardEvent) bool {
		return false
	},
	OnMouseEvent: func(event ViewportMouseEvent) bool {
		return false
	},
	OnRender: func(framebuffer render.Framebuffer, size ui.Size) {},
}

// ViewportMouseEvent is a wrapper on top of ui.MouseEvent which includes
// X and Y coordinates that are in the [0..1] range, making them size agnostic.
type ViewportMouseEvent struct {
	ui.MouseEvent
	X float32
	Y float32
}

// Viewport represents a component that uses low-level API calls to render
// an external scene, unlike other components that rely on the Canvas API.
var Viewport = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	essence := co.UseLifecycle(func(handle co.LifecycleHandle) *viewportEssence {
		return &viewportEssence{}
	})
	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence:   essence,
			Focusable: optional.Value(true),
		})
		co.WithLayoutData(props.LayoutData())
	})
})

var _ ui.Surface = (*viewportEssence)(nil)
var _ ui.ElementKeyboardHandler = (*viewportEssence)(nil)
var _ ui.ElementMouseHandler = (*viewportEssence)(nil)
var _ ui.ElementRenderHandler = (*viewportEssence)(nil)

type viewportEssence struct {
	co.BaseLifecycle

	api          render.API
	colorTexture render.Texture
	framebuffer  render.Framebuffer

	fbResizeBlocked bool
	fbDesiredWidth  int
	fbDesiredHeight int
	fbWidth         int
	fbHeight        int

	onKeyboardEvent func(event ui.KeyboardEvent) bool
	onMouseEvent    func(event ViewportMouseEvent) bool
	onRender        func(framebuffer render.Framebuffer, size ui.Size)
}

func (e *viewportEssence) OnCreate(props co.Properties, scope co.Scope) {
	e.OnUpdate(props, scope)
	e.createFramebuffer(ViewportInitialFBWidth, ViewportInitialFBHeight)
}

func (e *viewportEssence) OnUpdate(props co.Properties, scope co.Scope) {
	var (
		data         = co.GetData[ViewportData](props)
		callbackData = co.GetOptionalCallbackData(props, defaultViewportCallbackData)
	)
	e.api = data.API
	e.onKeyboardEvent = callbackData.OnKeyboardEvent
	if e.onKeyboardEvent == nil {
		e.onKeyboardEvent = defaultViewportCallbackData.OnKeyboardEvent
	}
	e.onMouseEvent = callbackData.OnMouseEvent
	if e.onMouseEvent == nil {
		e.onMouseEvent = defaultViewportCallbackData.OnMouseEvent
	}
	e.onRender = callbackData.OnRender
}

func (e *viewportEssence) OnDestroy(scope co.Scope) {
	e.releaseFramebuffer()
}

func (e *viewportEssence) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	return e.onKeyboardEvent(event)
}

func (e *viewportEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	size := element.Bounds().Size
	x := -1.0 + 2.0*float32(event.Position.X)/float32(size.Width)
	y := 1.0 - 2.0*float32(event.Position.Y)/float32(size.Height)
	return e.onMouseEvent(ViewportMouseEvent{
		MouseEvent: event,
		X:          x,
		Y:          y,
	})
}

func (e *viewportEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	size := element.Bounds().Size
	e.fbDesiredWidth = size.Width
	e.fbDesiredHeight = size.Height

	if e.fbDesiredWidth != e.fbWidth || e.fbDesiredHeight != e.fbHeight {
		if !e.fbResizeBlocked {
			e.releaseFramebuffer()
			e.createFramebuffer(e.fbDesiredWidth, e.fbDesiredHeight)
			e.fbResizeBlocked = true

			time.AfterFunc(time.Second, func() {
				co.Schedule(func() {
					e.fbResizeBlocked = false
					element.Invalidate()
				})
			})
		}
	}

	width := float32(size.Width)
	height := float32(size.Height)

	if e.onRender != nil {
		canvas.DrawSurface(
			e,
			sprec.ZeroVec2(),
			sprec.NewVec2(width, height),
		)
	} else {
		canvas.Reset()
		canvas.Rectangle(
			sprec.ZeroVec2(),
			sprec.NewVec2(width, height),
		)
		canvas.Fill(ui.Fill{
			Rule:  ui.FillRuleSimple,
			Color: BackgroundColor,
		})
	}
}

func (e *viewportEssence) Render() (render.Texture, ui.Size) {
	fbSize := ui.NewSize(e.fbWidth, e.fbHeight)
	e.onRender(e.framebuffer, fbSize)
	return e.colorTexture, fbSize
}

func (e *viewportEssence) createFramebuffer(width, height int) {
	e.fbWidth = width
	e.fbHeight = height
	e.colorTexture = e.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           width,
		Height:          height,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
	})
	e.framebuffer = e.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			e.colorTexture,
		},
	})
}

func (e *viewportEssence) releaseFramebuffer() {
	if e.framebuffer != nil {
		e.framebuffer.Release()
		e.framebuffer = nil
	}
	if e.colorTexture != nil {
		e.colorTexture.Release()
		e.colorTexture = nil
	}
}
