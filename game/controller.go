package game

import (
	"time"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/graphics"
	"github.com/mokiat/lacking/input"
)

type WindowSize struct {
	Width  int
	Height int
}

type InitContext struct {
	WindowSize WindowSize
	GFXWorker  *async.Worker
}

type UpdateContext struct {
	ElapsedTime time.Duration
	WindowSize  WindowSize
	Keyboard    *input.Keyboard
	Gamepad     *input.Gamepad
	GFXWorker   *async.Worker
}

type RenderContext struct {
	WindowSize  WindowSize
	GFXPipeline *graphics.Pipeline
}

type ReleaseContext struct {
	GFXWorker *async.Worker
}

type Controller interface {
	Init(InitContext) error
	Update(UpdateContext) bool
	Render(RenderContext)
	Release(ReleaseContext) error
}
