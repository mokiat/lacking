package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

// ContainerData holds the data for a Container component.
type ContainerData struct {
	BackgroundColor opt.T[ui.Color]
	BorderColor     opt.T[ui.Color]
	BorderSize      ui.Spacing
	Padding         ui.Spacing
	Layout          ui.Layout
}

var containerDefaultData = ContainerData{
	Layout: layout.Fill(),
}

// Container represents a component that holds other components and has
// some sort of visual boundary.
var Container = co.Define(&containerComponent{})

type containerComponent struct {
	co.BaseComponent

	backgroundColor ui.Color
	borderColor     ui.Color
	borderSize      ui.Spacing
	padding         ui.Spacing
	layout          ui.Layout
}

func (c *containerComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), containerDefaultData)
	if data.BackgroundColor.Specified {
		c.backgroundColor = data.BackgroundColor.Value
	} else {
		c.backgroundColor = SurfaceColor
	}
	if data.BorderColor.Specified {
		c.borderColor = data.BorderColor.Value
	} else {
		c.borderColor = OutlineColor
	}
	c.borderSize = data.BorderSize
	c.padding = data.Padding
	c.layout = data.Layout
}

func (c *containerComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence: c,
			Padding: c.padding,
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
	if !c.borderColor.Transparent() {
		canvas.Reset()
		canvas.SetStrokeColor(c.borderColor)
		if c.borderSize.Top > 0 {
			canvas.SetStrokeSizeSeparate(float32(c.borderSize.Top), 0.0)
			canvas.MoveTo(sprec.NewVec2(drawBounds.Width(), 0.0))
			canvas.LineTo(sprec.NewVec2(0.0, 0.0))
		}
		if c.borderSize.Left > 0 {
			canvas.SetStrokeSizeSeparate(float32(c.borderSize.Left), 0.0)
			canvas.MoveTo(sprec.NewVec2(0.0, 0.0))
			canvas.LineTo(sprec.NewVec2(0.0, drawBounds.Height()))
		}
		if c.borderSize.Bottom > 0 {
			canvas.SetStrokeSizeSeparate(float32(c.borderSize.Bottom), 0.0)
			canvas.MoveTo(sprec.NewVec2(0.0, drawBounds.Height()))
			canvas.LineTo(sprec.NewVec2(drawBounds.Width(), drawBounds.Height()))
		}
		if c.borderSize.Right > 0 {
			canvas.SetStrokeSizeSeparate(float32(c.borderSize.Right), 0.0)
			canvas.MoveTo(sprec.NewVec2(drawBounds.Width(), drawBounds.Height()))
			canvas.LineTo(sprec.NewVec2(drawBounds.Width(), 0.0))
		}
		canvas.Stroke()
	}
}
