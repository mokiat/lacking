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
var Paper = co.DefineType(&PaperComponent{})

type PaperComponent struct {
	Properties co.Properties `co:"properties"`

	layout ui.Layout
}

func (c *PaperComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, paperDefaultData)
	c.layout = data.Layout
}

func (c *PaperComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties.LayoutData())
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
		co.WithChildren(c.Properties.Children())
	})
}

func (c *PaperComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	bounds := element.Bounds()
	size := sprec.NewVec2(
		float32(bounds.Width),
		float32(bounds.Height),
	)

	canvas.Reset()
	canvas.SetStrokeSize(PaperBorderSize)
	canvas.SetStrokeColor(OutlineColor)
	canvas.RoundRectangle(
		sprec.ZeroVec2(),
		size,
		sprec.NewVec4(PaperRoundness, PaperRoundness, PaperRoundness, PaperRoundness),
	)
	canvas.Fill(ui.Fill{
		Color: SurfaceColor,
	})
	canvas.Stroke()
}
