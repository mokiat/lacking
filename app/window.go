package app

import "fmt"

type Window interface {
	SetTitle(title string)
	Resize(width, height int)
	Redraw()
	Close()
}

type WindowChangeHandler interface {
	OnWindowCreate(w Window)
	OnWindowDestroy(w Window)
	OnWindowFramebufferResize(w Window, width, height int)
	OnWindowUpdate(w Window)
	OnWindowCloseRequested(w Window)
}

type DefaultWindowChangeHandler struct{}

var _ WindowChangeHandler = DefaultWindowChangeHandler{}

func (DefaultWindowChangeHandler) OnWindowCreate(w Window) {
	w.Resize(512, 256)
}

func (DefaultWindowChangeHandler) OnWindowDestroy(w Window) {
}

func (DefaultWindowChangeHandler) OnWindowFramebufferResize(w Window, width, height int) {
	w.SetTitle(fmt.Sprintf("FB SIZE: %d / %d", width, height))
}

func (DefaultWindowChangeHandler) OnWindowUpdate(w Window) {
}

func (DefaultWindowChangeHandler) OnWindowCloseRequested(w Window) {
	w.Close()
}
