package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
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

var defaultPaperData = PaperData{
	Layout: NewFillLayout(),
}

// Paper represents an outlined container.
var Paper = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	data := co.GetOptionalData(props, defaultPaperData)

	essenceState := co.UseState(func() *paperEssence {
		return &paperEssence{}
	})

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essenceState.Get(),
			Layout:  data.Layout,
			Padding: ui.Spacing{
				Left:   PaperPadding,
				Right:  PaperPadding,
				Top:    PaperPadding,
				Bottom: PaperPadding,
			},
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementRenderHandler = (*paperEssence)(nil)

type paperEssence struct{}

func (e *paperEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
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
