package std

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	ToolbarHeight      = 48
	ToolbarBorderSize  = 1
	ToolbarSidePadding = 5
	ToolbarItemSpacing = 10
	ToolbarItemHeight  = ToolbarHeight - 2*ToolbarBorderSize
)

// ToolbarPositioning determines the visual appearance of the toolbar,
// depending on where it is intended to be placed on the screen.
type ToolbarPositioning int

const (
	ToolbarPositioningTop ToolbarPositioning = iota
	ToolbarPositioningMiddle
	ToolbarPositioningBottom
)

// ToolbarData can be used to specify configuration for the Toolbar component.
type ToolbarData struct {
	Positioning ToolbarPositioning
}

var toolbarDefaultData = ToolbarData{
	Positioning: ToolbarPositioningTop,
}

// Toolbar represents a container that holds key controls (mostly buttons)
// in a horizontal fashion.
var Toolbar = co.Define(&toolbarComponent{})

type toolbarComponent struct {
	co.BaseComponent

	positioning ToolbarPositioning
}

func (c *toolbarComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), toolbarDefaultData)
	c.positioning = data.Positioning
}

func (c *toolbarComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	drawBounds := canvas.DrawBounds(element, false)

	canvas.Reset()
	canvas.Rectangle(
		drawBounds.Position,
		drawBounds.Size,
	)
	canvas.Fill(ui.Fill{
		Color: SurfaceColor,
	})

	canvas.Reset()
	canvas.SetStrokeSize(float32(ToolbarBorderSize))
	canvas.SetStrokeColor(OutlineColor)
	if c.positioning != ToolbarPositioningTop {
		canvas.MoveTo(sprec.NewVec2(drawBounds.Width(), 0.0))
		canvas.LineTo(sprec.NewVec2(0.0, 0.0))
	}
	if c.positioning != ToolbarPositioningBottom {
		canvas.MoveTo(sprec.NewVec2(0.0, drawBounds.Height()))
		canvas.LineTo(sprec.NewVec2(drawBounds.Width(), drawBounds.Height()))
	}
	canvas.Stroke()
}

func (c *toolbarComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence: c,
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
		co.WithChildren(c.Properties().Children())
	})
}
