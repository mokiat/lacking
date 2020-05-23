package game

import (
	"time"

	"github.com/mokiat/lacking/graphics"
	"github.com/mokiat/lacking/input"
)

type WindowSize struct {
	Width  int
	Height int
}

type InitContext struct {
	WindowSize WindowSize
	GFXWorker  *graphics.Worker
}

type UpdateContext struct {
	ElapsedTime time.Duration
	WindowSize  WindowSize
	Keyboard    *input.Keyboard
	Gamepad     *input.Gamepad
	GFXWorker   *graphics.Worker
}

type RenderContext struct {
	WindowSize  WindowSize
	GFXPipeline *graphics.Pipeline
}

type ReleaseContext struct {
	GFXWorker *graphics.Worker
}

type Controller interface {
	Init(InitContext) error
	Update(UpdateContext) bool
	Render(RenderContext)
	Release(ReleaseContext) error
}
