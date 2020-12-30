package glfwapp

import (
	"fmt"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/mokiat/lacking/app"
)

func NewWindowConfig(title string, width, height int) *WindowConfig {
	return &WindowConfig{
		title:        title,
		width:        width,
		height:       height,
		swapInterval: 1,
	}
}

type WindowConfig struct {
	title        string
	width        int
	height       int
	swapInterval int
}

func (c *WindowConfig) SetVSync(vsync bool) {
	if vsync {
		c.swapInterval = 1
	} else {
		c.swapInterval = 0
	}
}

func CreateWindow(cfg *WindowConfig, handler app.WindowChangeHandler) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := glfw.Init(); err != nil {
		return fmt.Errorf("failed to initialize glfw: %w", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
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

	wrapperWindow := &Window{
		window:  window,
		handler: handler,
		actions: make(chan func(), 1024),
	}
	return wrapperWindow.loop()
}

var _ app.Window = (*Window)(nil)

type Window struct {
	window      *glfw.Window
	handler     app.WindowChangeHandler
	actions     chan func()
	shouldClose bool
}

func (w *Window) loop() error {
	w.handler.OnWindowCreate(w)
	defer w.handler.OnWindowDestroy(w)

	w.window.SetFramebufferSizeCallback(func(glfwWin *glfw.Window, width int, height int) {
		w.handler.OnWindowFramebufferResize(w, width, height)
	})

	w.window.SetRefreshCallback(func(*glfw.Window) {
		w.swapBuffers()
	})

	for !w.shouldClose {
		glfw.WaitEvents()

		if w.window.ShouldClose() {
			w.handler.OnWindowCloseRequested(w)
			w.window.SetShouldClose(false)
		}

		for action, ok := w.popAction(); ok; action, ok = w.popAction() {
			action()
		}
	}

	return nil
}

func (w *Window) pushSyncAction(action func()) {
	done := make(chan struct{})
	w.actions <- func() {
		action()
		close(done)
	}
	glfw.PostEmptyEvent()
	<-done
}

func (w *Window) pushAsyncAction(action func()) {
	w.actions <- action
	glfw.PostEmptyEvent()
}

func (w *Window) popAction() (func(), bool) {
	select {
	case action, ok := <-w.actions:
		return action, ok
	default:
		return nil, false
	}
}

func (w *Window) swapBuffers() {
	w.window.SwapBuffers()
}

func (w *Window) SetTitle(title string) {
	w.pushAsyncAction(func() {
		w.window.SetTitle(title)
	})
}

func (w *Window) Resize(width, height int) {
	w.pushAsyncAction(func() {
		w.window.SetSize(width, height)
	})
}

func (w *Window) Redraw() {
	w.pushSyncAction(func() {
		w.swapBuffers()
	})
}

func (w *Window) Close() {
	w.pushAsyncAction(func() {
		w.shouldClose = true
	})
}
