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
}

func (d *driver) Run() error {
	d.window.SetFramebufferSizeCallback(func(w *glfw.Window, width int, height int) {
		d.subscriber.OnContentResize(d, ui.NewSize(width, height))
	})
	fbSize := ui.NewSize(d.window.GetFramebufferSize())
	d.subscriber.OnContentResize(d, fbSize)

	d.window.SetRefreshCallback(func(w *glfw.Window) {
		d.renderContent()
	})

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

func (d *driver) renderContent() {
	d.subscriber.OnRender(d, nil) // TODO: Pass a canvas
	d.window.SwapBuffers()
}
