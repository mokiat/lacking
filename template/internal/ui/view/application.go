package view

import (
	"github.com/mokiat/lacking/template/internal/ui/model"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mvc"
	"github.com/mokiat/lacking/ui/std"
)

var Application = mvc.EventListener(co.Define(&applicationComponent{}))

type applicationComponent struct {
	co.BaseComponent

	appModel *model.Application
}

func (c *applicationComponent) OnCreate() {
	eventBus := co.TypedValue[*mvc.EventBus](c.Scope())
	c.appModel = model.NewApplication(eventBus)
}

func (c *applicationComponent) Render() co.Instance {
	return co.New(std.Switch, func() {
		co.WithData(std.SwitchData{
			ChildKey: c.appModel.ActiveView(),
		})

		co.WithChild(model.ViewNameIntro, co.New(IntroScreen, func() {
			co.WithData(IntroScreenData{
				AppModel: c.appModel,
			})
		}))
		co.WithChild(model.ViewNamePlay, co.New(PlayScreen, func() {
			co.WithData(PlayScreenData{
				AppModel: c.appModel,
			})
		}))
	})
}

func (c *applicationComponent) OnEvent(event mvc.Event) {
	switch event.(type) {
	case model.ApplicationActiveViewChangedEvent:
		c.Invalidate()
	}
}
