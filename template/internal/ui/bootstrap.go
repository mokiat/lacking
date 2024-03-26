package ui

import (
	"github.com/mokiat/lacking/template/internal/ui/view"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mvc"
)

func BootstrapApplication(window *ui.Window) {
	eventBus := mvc.NewEventBus()

	scope := co.RootScope(window)
	scope = co.TypedValueScope(scope, eventBus)
	co.Initialize(scope, co.New(Bootstrap, nil))
}

var Bootstrap = co.Define(&bootstrapComponent{})

type bootstrapComponent struct {
	co.BaseComponent
}

func (c *bootstrapComponent) Render() co.Instance {
	return co.New(view.Application, nil)
}
