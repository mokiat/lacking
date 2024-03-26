package view

import (
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/template/internal/ui/model"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

type IntroScreenData struct {
	AppModel *model.Application
}

var IntroScreen = co.Define(&introScreenComponent{})

type introScreenComponent struct {
	co.BaseComponent
}

func (c *introScreenComponent) OnCreate() {
	co.Window(c.Scope()).SetCursorVisible(false)

	screenData := co.GetData[IntroScreenData](c.Properties())
	appModel := screenData.AppModel

	co.After(c.Scope(), time.Second, func() {
		appModel.SetActiveView(model.ViewNamePlay)
	})
}

func (c *introScreenComponent) OnDelete() {
	co.Window(c.Scope()).SetCursorVisible(true)
}

func (c *introScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("logo-picture", co.New(std.Picture, func() {
			co.WithLayoutData(layout.Data{
				Width:            opt.V(512),
				Height:           opt.V(128),
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
			co.WithData(std.PictureData{
				BackgroundColor: opt.V(ui.Transparent()),
				Image:           co.OpenImage(c.Scope(), "ui/images/logo.png"),
				Mode:            std.ImageModeFit,
			})
		}))
	})
}
