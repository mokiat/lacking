package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

var (
	ToolbarSeparatorWidth           = 3
	ToolbarSeparatorLineSize        = float32(1.0)
	ToolbarSeparatorLineLengthRatio = float32(0.7)
)

var ToolbarSeparator = co.Define(&toolbarSeparatorComponent{})

type toolbarSeparatorComponent struct {
	co.BaseComponent
}

func (c *toolbarSeparatorComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence: c,
			IdealSize: opt.V(ui.NewSize(
				ToolbarSeparatorWidth,
				ToolbarItemHeight,
			)),
		})
	})
}

func (c *toolbarSeparatorComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	drawBounds := canvas.DrawBounds(element, false)
	halfWidth := drawBounds.Width() / 2.0
	lineLength := drawBounds.Height() * ToolbarSeparatorLineLengthRatio
	linePadding := (drawBounds.Height() - lineLength) / 2.0

	canvas.Reset()
	canvas.SetStrokeSizeSeparate(
		ToolbarSeparatorLineSize/2.0,
		ToolbarSeparatorLineSize/2.0,
	)
	canvas.SetStrokeColor(OutlineColor)
	canvas.MoveTo(
		sprec.NewVec2(halfWidth, float32(linePadding)),
	)
	canvas.LineTo(
		sprec.NewVec2(halfWidth, float32(linePadding+lineLength)),
	)
	canvas.Stroke()
}
