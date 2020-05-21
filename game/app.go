package game

import (
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func NewApp(cfg AppConfig) *App {
	return &App{
		cfg: cfg,
	}
}

type App struct {
	cfg AppConfig
}

func (a *App) Run(controller Controller) error {
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

	window, err := glfw.CreateWindow(a.cfg.WindowWidth, a.cfg.WindowHeight, a.cfg.WindowTitle, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create glfw window: %w", err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()
	if a.cfg.WindowHideCursor {
		window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	}

	if a.cfg.WindowVSync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}

	if err := gl.Init(); err != nil {
		return fmt.Errorf("failed to initialize opengl: %w", err)
	}

	loop := &updateLoop{
		controller: controller,
		interval:   a.cfg.UpdateLoopInterval,
		stop:       make(chan struct{}),
		finished:   make(chan struct{}),
	}

	window.SetFramebufferSizeCallback(func(w *glfw.Window, width int, height int) {
		loop.SetWindowSize(WindowSize{
			Width:  width,
			Height: height,
		})
	})
	fbWidth, fbHeight := window.GetFramebufferSize()
	loop.SetWindowSize(WindowSize{
		Width:  fbWidth,
		Height: fbHeight,
	})

	go loop.Run()

	for !window.ShouldClose() && loop.IsRunning() {
		// TODO: Remove; replace with renderer
		gl.ClearColor(0.3, 0.6, 1.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		window.SwapBuffers()
		glfw.PollEvents()
	}

	loop.Stop()

	// TODO: GL cleanup

	return nil
}

type updateLoop struct {
	controller Controller
	interval   time.Duration
	stop       chan struct{}
	finished   chan struct{}
	windowSize atomic.Value
}

func (l *updateLoop) SetWindowSize(size WindowSize) {
	l.windowSize.Store(size)
}

func (l *updateLoop) IsRunning() bool {
	select {
	case <-l.finished:
		return false
	default:
		return true
	}
}

func (l *updateLoop) Run() {
	defer close(l.finished)

	initCtx := InitContext{
		WindowSize: l.windowSize.Load().(WindowSize),
	}
	if err := l.controller.Init(initCtx); err != nil {
		log.Printf("controller init error: %v", err)
		return
	}

	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	lastTick := time.Now()
	running := true
	for running {
		select {
		case currentTick := <-ticker.C:
			updateCtx := UpdateContext{
				ElapsedTime: currentTick.Sub(lastTick),
				WindowSize:  l.windowSize.Load().(WindowSize),
			}
			running = l.controller.Update(updateCtx)
			lastTick = currentTick
		case <-l.stop:
			running = false
		}
	}

	releaseCtx := ReleaseContext{}
	if err := l.controller.Release(releaseCtx); err != nil {
		log.Printf("controller release error: %v", err)
		return
	}
}

func (l *updateLoop) Stop() {
	close(l.stop)
	<-l.finished
}
