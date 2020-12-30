package glui

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/ui"
)

func NewManager() *Manager {
	return &Manager{}
}

var _ ui.Manager = (*Manager)(nil)
var _ app.WindowChangeHandler = (*Manager)(nil)

type Manager struct {
	app.DefaultWindowChangeHandler
}
