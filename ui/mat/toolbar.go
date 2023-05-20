package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	ToolbarHeight      = 64
	ToolbarBorderSize  = 1
	ToolbarSidePadding = 5
	ToolbarItemSpacing = 10
	ToolbarItemHeight  = ToolbarHeight - 2*ToolbarBorderSize
)

const (
	ToolbarOrientationLeftToRight ToolbarOrientation = iota
	ToolbarOrientationRightToLeft
)

// ToolbarOrientation determines the direction in which child components
// are laid out.
type ToolbarOrientation int

const (
	ToolbarPositioningTop ToolbarPositioning = iota
	ToolbarPositioningMiddle
	ToolbarPositioningBottom
)

// ToolbarPositioning determines the visual appearance of the toolbar,
// depending on where it is intended to be placed on the screen.
type ToolbarPositioning int

// ToolbarData can be used to specify configuration for the Toolbar component.
type ToolbarData struct {
	Orientation ToolbarOrientation
	Positioning ToolbarPositioning
}

var defaultToolbarData = ToolbarData{}

// Toolbar represents a container that holds key controls (mostly buttons)
// in a horizontal fashion.
var Toolbar = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data       = co.GetOptionalData(props, defaultToolbarData)
		layoutData = co.GetOptionalLayoutData(props, layout.Data{})
	)

	essence := co.UseState(func() *toolbarEssence {
		return &toolbarEssence{}
	}).Get()
	essence.positioning = data.Positioning

	// force specific height
	layoutData.Height = opt.V(ToolbarHeight)

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Padding: ui.Spacing{
				Left:   ToolbarSidePadding,
				Right:  ToolbarSidePadding,
				Top:    ToolbarBorderSize,
				Bottom: ToolbarBorderSize,
			},
			Layout: layout.Horizontal(layout.HorizontalSettings{
				ContentAlignment: layout.VerticalAlignmentCenter,
				ContentSpacing:   ToolbarItemSpacing,
			}),
		})
		co.WithLayoutData(layoutData)
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementRenderHandler = (*toolbarEssence)(nil)

type toolbarEssence struct {
	positioning ToolbarPositioning
}

func (e *toolbarEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	bounds := element.Bounds()
	size := sprec.NewVec2(
		float32(bounds.Width),
		float32(bounds.Height),
	)

	canvas.Reset()
	canvas.Rectangle(
		sprec.ZeroVec2(),
		size,
	)
	canvas.Fill(ui.Fill{
		Color: SurfaceColor,
	})

	canvas.Reset()
	canvas.SetStrokeSize(float32(ToolbarBorderSize))
	canvas.SetStrokeColor(OutlineColor)
	if e.positioning != ToolbarPositioningTop {
		canvas.MoveTo(sprec.NewVec2(size.X, 0.0))
		canvas.LineTo(sprec.NewVec2(0.0, 0.0))
	}
	if e.positioning != ToolbarPositioningBottom {
		canvas.MoveTo(sprec.NewVec2(0.0, size.Y))
		canvas.LineTo(sprec.NewVec2(size.X, size.Y))
	}
	canvas.Stroke()
}
