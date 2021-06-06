package graphics

import (
	"github.com/mokiat/lacking/game/graphics"
)

func NewEngine() *Engine {
	return &Engine{
		renderer: newRenderer(),
	}
}

var _ graphics.Engine = (*Engine)(nil)

type Engine struct {
	renderer *Renderer
}

func (e *Engine) Create() {
	e.renderer.Allocate()
}

func (e *Engine) CreateScene() graphics.Scene {
	return newScene(e.renderer)
}

func (e *Engine) Destroy() {
	e.renderer.Release()
}
