package game

import "time"

type WindowSize struct {
	Width  int
	Height int
}

type InitContext struct {
	WindowSize WindowSize
}

type UpdateContext struct {
	ElapsedTime time.Duration
	WindowSize  WindowSize
}

type ReleaseContext struct {
}

type Controller interface {
	Init(InitContext) error
	Update(UpdateContext) bool
	Release(ReleaseContext) error
}
