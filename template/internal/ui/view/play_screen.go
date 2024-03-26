package view

import (
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/debug/metric/metricui"
	"github.com/mokiat/lacking/template/internal/ui/model"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
)

var PlayScreen = co.Define(&playScreenComponent{})

type PlayScreenData struct {
	AppModel *model.Application
}

type playScreenComponent struct {
	co.BaseComponent

	appModel *model.Application

	debugVisible bool
}

var _ ui.ElementKeyboardHandler = (*playScreenComponent)(nil)

func (c *playScreenComponent) OnCreate() {
	data := co.GetData[PlayScreenData](c.Properties())
	c.appModel = data.AppModel

	c.debugVisible = false
}

func (c *playScreenComponent) OnDelete() {
}

func (c *playScreenComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Code {

	case ui.KeyCodeEscape:
		co.Window(c.Scope()).Close()
		return true

	case ui.KeyCodeTab:
		if event.Action == ui.KeyboardActionDown {
			c.debugVisible = !c.debugVisible
			c.Invalidate()
		}
		return true

	default:
		return false
	}
}

func (c *playScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.RGB(0x11, 0x22, 0x33)),
			Layout:          layout.Fill(),
		})

		co.WithChild("root", co.New(std.Element, func() {
			co.WithData(std.ElementData{
				Essence:   c,
				Focusable: opt.V(true),
				Focused:   opt.V(true),
				Layout:    layout.Anchor(),
			})

			if c.debugVisible {
				co.WithChild("flamegraph", co.New(metricui.FlameGraph, func() {
					co.WithData(metricui.FlameGraphData{
						UpdateInterval: time.Second,
					})
					co.WithLayoutData(layout.Data{
						Top:   opt.V(0),
						Left:  opt.V(0),
						Right: opt.V(0),
					})
				}))
			}

			co.WithChild("label", co.New(std.Label, func() {
				co.WithLayoutData(layout.Data{
					HorizontalCenter: opt.V(0),
					VerticalCenter:   opt.V(0),
				})
				co.WithData(std.LabelData{
					Font:      co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf"),
					FontSize:  opt.V(float32(64.0)),
					FontColor: opt.V(ui.White()),
					Text:      "Your game would go here...",
				})
			}))
		}))
	})
}
