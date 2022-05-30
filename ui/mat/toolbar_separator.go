package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

var (
	ToolbarSeparatorWidth           = 3
	ToolbarSeparatorLineSize        = float32(1.0)
	ToolbarSeparatorLineLengthRatio = float32(0.7)
)

// ToolbarSeparator separates controls within a Toolbar container.
// It is visualized as a vertical line.
var ToolbarSeparator = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	essence := co.UseState(func() *toolbarSeparatorEssence {
		return &toolbarSeparatorEssence{}
	}).Get()

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			IdealSize: optional.Value(ui.NewSize(
				ToolbarSeparatorWidth,
				ToolbarItemHeight,
			)),
		})
		co.WithLayoutData(props.LayoutData())
	})
})

var _ ui.ElementRenderHandler = (*toolbarSeparatorEssence)(nil)

type toolbarSeparatorEssence struct{}

func (e *toolbarSeparatorEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	bounds := element.Bounds()
	size := sprec.NewVec2(
		float32(bounds.Width),
		float32(bounds.Height),
	)
	halfWidth := size.X / 2.0
	lineLength := size.Y * ToolbarSeparatorLineLengthRatio
	linePadding := (size.Y - lineLength) / 2.0

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
