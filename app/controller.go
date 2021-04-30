package app

// Controller is a mechanism through which the user code can be
// notified of changes to the application window.
//
// All methods will be invoked on the UI thread (UI goroutine),
// unless specified otherwise.
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
	// especially the case on devices with high DPI setting.
	OnFramebufferResize(window Window, width, height int)

	// OnKeyboardEvent is called whenever a keyboard event has occurred.
	OnKeyboardEvent(window Window, event KeyboardEvent) bool

	// OnMouseEvent is called whenever a mouse event has occurred.
	OnMouseEvent(window Window, event MouseEvent) bool

	// OnRender is called whenever the window would like to be redrawn.
	OnRender(window Window)

	// OnCloseRequested is called whenever the end-user has requested
	// that the application be closed through the native OS means
	// (e.g. pressing the close button or ALT+F4).
	//
	// It is up to the controller implementation to call Close on the
	// window if they accept the end-user's request (e.g. games might
	// decide to show a prompt and not want the close to occur).
	OnCloseRequested(window Window)

	// OnDestroy is called whenever the window is about to close.
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

func (NopController) OnCloseRequested(window Window) {}

func (NopController) OnDestroy(window Window) {}

// NewLayeredController returns a new LayeredController that has
// the specified layers configured.
func NewLayeredController(layers ...Controller) *LayeredController {
	return &LayeredController{
		layers: layers,
	}
}

var _ (Controller) = (*LayeredController)(nil)

// LayeredController is an implementation of Controller that invokes
// the specified controller layers in a certain order, mostly favouring
// layers with higher index.
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

func (c *LayeredController) OnCloseRequested(window Window) {
	for i := 0; i < len(c.layers); i++ {
		c.layers[i].OnCloseRequested(window)
	}
}

func (c *LayeredController) OnDestroy(window Window) {
	for i := len(c.layers) - 1; i >= 0; i-- {
		c.layers[i].OnDestroy(window)
	}
}
