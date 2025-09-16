---
title: Overview
---

# User Interface

The user interface API of lacking is comprised of multiple layers that are more commonly seen in standard desktop or web UI, instead of gaming. This has some historical reasons but the idea was also to have something that can be used for app/tool development as well.

## Element API

The core layer of the API represents the user interface in a similar way to web pages. A window is comprised of a number of nested Elements that each can have custom rendering behavior and input event handling.

The way Elements are created is imperative, which can be more optimal and reduce memory usage but requires more boilerplate and manual work to coordinate, especially when certain UI elements need to disappear and/or be replaced with something else.

Following is a rough idea how this can be used.

```go
// initUI function can be passed to the UI controller bootstrap function.
func initUI(window *ui.Window) {
	container := NewContainer(window.Root())
	container.SetLayout(layout.Anchor()) // use anchor layout
	container.SetBackgroundColor(ui.Navy())

	label := NewLabel(container.Element())
	label.SetText("Hello World")
	label.SetLayoutConfig(layout.Data{ // position in the center of the container
		HorizontalCenter: opt.V(0),
		VerticalCenter:   opt.V(0),
	})
}
```

A container component can then be implemented as follows.

```go
import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/layout"
)

func NewContainer(parentElement *ui.Element) *Container {
	element := parentElement.Window().CreateElement()
	result := &Container{
		element:         element,
		backgroundColor: ui.Black(),
	}
	element.SetEssence(result)
	parentElement.AppendChild(element)
	return result
}

type Container struct {
	element         *ui.Element
	backgroundColor ui.Color
}

var _ ui.ElementRenderHandler = (*Container)(nil)

func (c *Container) Element() *ui.Element {
	return c.element
}

func (c *Container) SetLayout(layout ui.Layout) {
	c.element.SetLayout(layout)
	c.element.Invalidate()
}

func (c *Container) SetLayoutConfig(config layout.Data) {
	c.element.SetLayoutConfig(config)
	c.element.Invalidate()
}

func (c *Container) SetBackgroundColor(color ui.Color) {
	c.backgroundColor = color
	c.element.Invalidate()
}

func (c *Container) OnRender(element *ui.Element, canvas *ui.Canvas) {
	bounds := canvas.DrawBounds(element, false)
	canvas.Push()
	canvas.Translate(bounds.Position)
	canvas.Reset() // prepare for new shape
	canvas.Rectangle(sprec.ZeroVec2(), bounds.Size)
	canvas.Fill(ui.Fill{
		Color: c.backgroundColor,
	})
	canvas.Pop()
}
```

And a label can be implemented as follows.

```go
import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/layout"
)

func NewLabel(parentElement *ui.Element) *Label {
	element := parentElement.Window().CreateElement()
	font, _ := parentElement.Window().Context().OpenFont("ui:///roboto-bold.ttf")
	result := &Label{
		element:  element,
		font:     font,
		fontSize: float32(24.0),
		text:     []rune("Label"),
	}
	result.updateIdealSize()
	element.SetEssence(result)
	parentElement.AppendChild(element)
	return result
}

type Label struct {
	element  *ui.Element
	font     *ui.Font
	fontSize float32
	text     []rune
}

var _ ui.ElementRenderHandler = (*Label)(nil)

func (l *Label) Element() *ui.Element {
	return l.element
}

func (l *Label) SetText(text string) {
	l.text = []rune(text)
	l.updateIdealSize()
}

func (l *Label) SetLayoutConfig(config layout.Data) {
	l.element.SetLayoutConfig(config)
	l.element.Invalidate()
}

func (l *Label) OnRender(element *ui.Element, canvas *ui.Canvas) {
	bounds := canvas.DrawBounds(element, false)
	canvas.Push()
	canvas.Translate(bounds.Position)
	textWidth := l.font.LineWidth(l.text, l.fontSize)
	textHeight := l.font.LineHeight(l.fontSize)
	textPosition := sprec.Vec2{
		X: (bounds.Size.X - textWidth) / 2.0,
		Y: (bounds.Size.Y - textHeight) / 2.0,
	}
	canvas.FillTextLine(l.text, textPosition, ui.Typography{
		Font:  l.font,
		Size:  l.fontSize,
		Color: ui.White(),
	})
	canvas.Pop()
}

func (l *Label) updateIdealSize() {
	textWidth := l.font.LineWidth(l.text, l.fontSize)
	textHeight := l.font.LineHeight(l.fontSize)
	l.element.SetIdealSize(ui.Size{
		Width:  int(sprec.Ceil(textWidth)),
		Height: int(sprec.Ceil(textHeight)),
	})
}
```

## Component API

Using the Element API as shown above it is perfectly possible to construct a complete app UI but has some downsides especially when having to deal with dynamically adding and removing children.

As such, the lacking framework includes a higher-level API that is declarative in nature. It is heavily inspired by frameworks like React, Vue, Svelte and similar.

Building a UI page like in the Element example would look as follows.


```go
// initUI function can be passed to the UI controller bootstrap function.
func initUI(window *ui.Window) {
	scope := co.RootScope(window)
	co.Initialize(scope, co.New(App, nil))
}

var App = co.Define[*appComponent]()

type appComponent struct {
	co.BaseComponent
}

func (c *appComponent) Render() co.Instance {
	return co.New(Container, func() {
		co.WithData(ContainerData{
			BackgroundColor: opt.V(ui.Navy()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("label", co.New(Label, func() {
			co.WithLayoutData(layout.Data{
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
			co.WithData(LabelData{
				Font: co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf"),
				FontSize: opt.V(float32(24.0)),
				FontColor: opt.V(ui.White()),
				Text: "First Button",
			})
		}))
	})
}

```

A container component can be implemented as follows.

```go
var Container = co.Define[*containerComponent]()

type ContainerData struct {
	BackgroundColor opt.T[ui.Color]
}

type containerComponent struct {
	co.BaseComponent
	layout          ui.Layout
	backgroundColor ui.Color
}


func (c *containerComponent) OnUpsert() {
	data := co.GetData[ContainerData](c.Properties())
  c.layout = data.Layout
	if data.BackgroundColor.Specified {
		c.backgroundColor = data.BackgroundColor.Value
	} else {
		c.backgroundColor = ui.Black()
	}
}

func (c *containerComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence: c,
			Layout:  c.layout,
		})
		co.WithChildren(c.Properties().Children())
	})
}


func (c *containerComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	drawBounds := canvas.DrawBounds(element, false)
	if !c.backgroundColor.Transparent() {
		canvas.Reset()
		canvas.Rectangle(drawBounds.Position, drawBounds.Size)
		canvas.Fill(ui.Fill{
			Color: c.backgroundColor,
		})
	}
}
```

The lacking framework includes a package [std](https://pkg.go.dev/github.com/mokiat/lacking/ui/std) (short for standard) that includes some basic component implementations. While not too pretty and unlikely to be used in a game, they can be useful when creating a tool or getting started with the component API.
