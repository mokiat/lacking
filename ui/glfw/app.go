package glfw

import (
	"fmt"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/mokiat/lacking/ui"
)

type Bootstrap interface {
	Run(window ui.Window)
}

type BootstrapFunc func(window ui.Window)

func (f BootstrapFunc) Run(window ui.Window) {
	f(window)
}

func RunApplication(cfg *AppConfig, bootstrap Bootstrap) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := glfw.Init(); err != nil {
		return fmt.Errorf("failed to initialize glfw: %w", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.SRGBCapable, glfw.True)

	window, err := glfw.CreateWindow(cfg.width, cfg.height, cfg.title, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create glfw window: %w", err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()
	glfw.SwapInterval(cfg.swapInterval)

	uiDriver := &driver{
		window: window,
	}
	uiWindow, subscriber := ui.CreateWindow(uiDriver)
	uiDriver.subscriber = subscriber

	subscriber.OnCreate(uiDriver)
	defer subscriber.OnDestroy(uiDriver)

	bootstrap.Run(uiWindow)
	return uiDriver.Run()
}

type driver struct {
	window     *glfw.Window
	subscriber ui.DriverSubscriber
	shouldStop bool
	shouldDraw bool

	// On OSX we get mouse events even outside the window
	// so we should keep track.
	mouseInside bool
}

func (d *driver) Run() error {
	d.window.SetRefreshCallback(d.onGLFWRefresh)

	d.window.SetSizeCallback(d.onGLFWSize)
	d.window.SetFramebufferSizeCallback(d.onGLFWFramebufferSize)
	fbSize := ui.NewSize(d.window.GetFramebufferSize())
	d.onGLFWFramebufferSize(d.window, fbSize.Width, fbSize.Height)

	d.window.SetKeyCallback(d.onGLFWKey)
	d.window.SetCharCallback(d.onGLFWChar)
	d.window.SetCursorPosCallback(d.onGLFWCursorPos)
	d.window.SetCursorEnterCallback(d.onGLFWCursorEnter)

	for !d.shouldStop {
		glfw.WaitEvents()

		if d.window.ShouldClose() {
			d.subscriber.OnCloseRequested(d)
			d.window.SetShouldClose(false)
		}

		if d.shouldDraw {
			d.shouldDraw = false
			d.renderContent()
		}
	}
	return nil
}

func (d *driver) SetTitle(title string) {
	d.window.SetTitle(title)
}

func (d *driver) SetSize(size ui.Size) {
	d.window.SetSize(size.Width, size.Height)
}

func (d *driver) Size() ui.Size {
	width, height := d.window.GetSize()
	return ui.NewSize(width, height)
}

func (d *driver) Redraw() {
	d.shouldDraw = true
}

func (d *driver) Destroy() {
	d.shouldStop = true
}

func (d *driver) onGLFWRefresh(w *glfw.Window) {
	d.renderContent()
}

func (d *driver) onGLFWSize(w *glfw.Window, width int, height int) {
	d.subscriber.OnResize(d, ui.NewSize(width, height))
}

func (d *driver) onGLFWFramebufferSize(w *glfw.Window, width int, height int) {
	// TODO: Resize the canvas. The framebuffer could be twice the size
	// of the window but the ui window need not know, since this additional
	// size is used for finer image quality and should be handled by the
	// canvas. (It should set the Ortho to the Size while using a FramebufferSize
	// framebuffer.)
}

func (d *driver) onGLFWChar(w *glfw.Window, char rune) {
	d.subscriber.OnKeyboardEvent(d, ui.KeyboardEvent{
		Type: ui.KeyboardEventTypeType,
		Rune: char,
	})
}

func (d *driver) onGLFWKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	eventType, ok := typeMapping[action]
	if !ok {
		return
	}

	keyCode, ok := keyMapping[key]
	if !ok {
		return
	}

	var modifiers ui.KeyModifierSet
	if (mods & glfw.ModControl) == glfw.ModControl {
		modifiers = modifiers | ui.KeyModifierSet(ui.KeyModifierControl)
	}
	if (mods & glfw.ModShift) == glfw.ModShift {
		modifiers = modifiers | ui.KeyModifierSet(ui.KeyModifierShift)
	}
	if (mods & glfw.ModAlt) == glfw.ModAlt {
		modifiers = modifiers | ui.KeyModifierSet(ui.KeyModifierAlt)
	}
	if (mods & glfw.ModCapsLock) == glfw.ModCapsLock {
		modifiers = modifiers | ui.KeyModifierSet(ui.KeyModifierCapsLock)
	}

	d.subscriber.OnKeyboardEvent(d, ui.KeyboardEvent{
		Type:      eventType,
		Code:      keyCode,
		Modifiers: modifiers,
	})
}

func (d *driver) onGLFWCursorPos(w *glfw.Window, xpos float64, ypos float64) {
	if !d.mouseInside {
		return
	}
	d.subscriber.OnMouseEvent(d, ui.MouseEvent{
		Index: 0,
		X:     int(xpos),
		Y:     int(ypos),
		Type:  ui.MouseEventTypeMove,
	})
}

func (d *driver) onGLFWCursorEnter(w *glfw.Window, entered bool) {
	var eventType ui.MouseEventType
	d.mouseInside = entered
	if entered {
		eventType = ui.MouseEventTypeEnter
	} else {
		eventType = ui.MouseEventTypeLeave
	}
	xpos, ypos := d.window.GetCursorPos()
	d.subscriber.OnMouseEvent(d, ui.MouseEvent{
		Index: 0,
		X:     int(xpos),
		Y:     int(ypos),
		Type:  eventType,
	})
}

func (d *driver) renderContent() {
	d.subscriber.OnRender(d, nil) // TODO: Pass a canvas
	d.window.SwapBuffers()
}
