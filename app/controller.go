package app

// Controller is a mechanism through which the user code can be
// notified of changes to the application window.
//
// All methods will be invoked on the UI thread (UI goroutine),
// unless otherwise specified.
type Controller interface {

	// OnCreate is called when the window has been created and is
	// ready to be used.
	OnCreate(window Window)

	// OnResize is called when the window's content area has been
	// resized.
	OnResize(window Window, width, height int)

	// OnFramebufferResize is called when the window's framebuffer has
	// been resized.
	// Note that the framebuffer need not match the content size. This
	// is mostly the case on devices with a high DPI setting.
	OnFramebufferResize(window Window, width, height int)

	// OnKeyboardEvent is called whenever a keyboard event has occurred.
	//
	// Return true to indicate that the event has been consumed and should
	// not be propagated to other potential receivers, otherwise return false.
	OnKeyboardEvent(window Window, event KeyboardEvent) bool

	// OnMouseEvent is called whenever a mouse event has occurred.
	//
	// Return true to indicate that the event has been consumed and should
	// not be propagated to other potential receivers, otherwise return false.
	OnMouseEvent(window Window, event MouseEvent) bool

	// OnRender is called whenever the window would like to be redrawn.
	OnRender(window Window)

	// OnCloseRequested is called whenever the end-user has requested
	// that the application be closed through the native OS means
	// (e.g. pressing the close button or ALT+F4).
	//
	// Unlike regular events, returning false from this function indicates
	// that the operation should not be performed. Furthermore, the event will
	// not be passed to other potential receivers.
	//
	// Note that on some platforms (e.g. browser) returning false may invoke
	// a native warning dialog.
	OnCloseRequested(window Window) bool

	// OnDestroy is called before the window is closed.
	OnDestroy(window Window)
}

var _ (Controller) = (*NopController)(nil)

// NopController is a no-op implementation of a Controller.
type NopController struct{}

func (NopController) OnCreate(window Window) {}

func (NopController) OnResize(window Window, width, height int) {}

func (NopController) OnFramebufferResize(window Window, width, height int) {}

func (NopController) OnKeyboardEvent(window Window, event KeyboardEvent) bool { return false }

func (NopController) OnMouseEvent(window Window, event MouseEvent) bool { return false }

func (NopController) OnRender(window Window) {}

func (NopController) OnCloseRequested(window Window) bool { return true }

func (NopController) OnDestroy(window Window) {}

// NewLayeredController returns a new LayeredController that has
// the specified controller layers configured.
func NewLayeredController(layers ...Controller) *LayeredController {
	return &LayeredController{
		layers: layers,
	}
}

var _ (Controller) = (*LayeredController)(nil)

// LayeredController is an implementation of Controller that invokes
// the specified controller layers in an order emulating multiple overlays
// of a window.
type LayeredController struct {
	layers []Controller
}

func (c *LayeredController) OnCreate(window Window) {
	for i := 0; i < len(c.layers); i++ {
		c.layers[i].OnCreate(window)
	}
}

func (c *LayeredController) OnResize(window Window, width, height int) {
	for i := 0; i < len(c.layers); i++ {
		c.layers[i].OnResize(window, width, height)
	}
}

func (c *LayeredController) OnFramebufferResize(window Window, width, height int) {
	for i := 0; i < len(c.layers); i++ {
		c.layers[i].OnFramebufferResize(window, width, height)
	}
}

func (c *LayeredController) OnKeyboardEvent(window Window, event KeyboardEvent) bool {
	for i := len(c.layers) - 1; i >= 0; i-- {
		if c.layers[i].OnKeyboardEvent(window, event) {
			return true
		}
	}
	return false
}

func (c *LayeredController) OnMouseEvent(window Window, event MouseEvent) bool {
	for i := len(c.layers) - 1; i >= 0; i-- {
		if c.layers[i].OnMouseEvent(window, event) {
			return true
		}
	}
	return false
}

func (c *LayeredController) OnRender(window Window) {
	for i := 0; i < len(c.layers); i++ {
		c.layers[i].OnRender(window)
	}
}

func (c *LayeredController) OnCloseRequested(window Window) bool {
	for i := len(c.layers) - 1; i >= 0; i-- {
		if !c.layers[i].OnCloseRequested(window) {
			return false
		}
	}
	return true
}

func (c *LayeredController) OnDestroy(window Window) {
	for i := len(c.layers) - 1; i >= 0; i-- {
		c.layers[i].OnDestroy(window)
	}
}
