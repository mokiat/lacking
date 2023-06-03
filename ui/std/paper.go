package std

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	PaperPadding    = 5
	PaperBorderSize = float32(1.0)
	PaperRoundness  = float32(10.0)
)

// PaperData can be used to configure a Paper component.
type PaperData struct {
	Layout ui.Layout
}

var paperDefaultData = PaperData{
	Layout: layout.Fill(),
}

// Paper represents an outlined container.
var Paper = co.Define(&paperComponent{})

type paperComponent struct {
	co.BaseComponent

	layout ui.Layout
}

func (c *paperComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), paperDefaultData)
	c.layout = data.Layout
}

func (c *paperComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence: c,
			Layout:  c.layout,
			Padding: ui.Spacing{
				Left:   PaperPadding,
				Right:  PaperPadding,
				Top:    PaperPadding,
				Bottom: PaperPadding,
			},
		})
		co.WithChildren(c.Properties().Children())
	})
}

func (c *paperComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	drawBounds := canvas.DrawBounds(element, false)

	canvas.Reset()
	canvas.SetStrokeSize(PaperBorderSize)
	canvas.SetStrokeColor(OutlineColor)
	canvas.RoundRectangle(
		drawBounds.Position,
		drawBounds.Size,
		sprec.NewVec4(PaperRoundness, PaperRoundness, PaperRoundness, PaperRoundness),
	)
	canvas.Fill(ui.Fill{
		Color: SurfaceColor,
	})
	canvas.Stroke()
}
