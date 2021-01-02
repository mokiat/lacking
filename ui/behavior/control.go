package behavior

import (
	"github.com/mokiat/lacking/ui"
)

const (
	defaultControlWidth  = 64
	defaultControlHeight = 32
)

func BuildControl(ctx ui.BuildContext) *Control {
	return &Control{
		id: ctx.Template.ID(),
		bounds: ui.Bounds{
			X:      0,
			Y:      0,
			Width:  defaultControlWidth,
			Height: defaultControlHeight,
		},
		layoutData: ctx.LayoutData,
	}
}

type Control struct {
	id         string
	layoutData ui.LayoutData
	bounds     ui.Bounds
}

var _ ui.Control = (*Control)(nil)

func (c *Control) ID() string {
	return c.id
}

func (c *Control) LayoutData() ui.LayoutData {
	return c.layoutData
}

func (c *Control) SetBounds(bounds ui.Bounds) {
	c.bounds = bounds
}

func (c *Control) Bounds() ui.Bounds {
	return c.bounds
}

func (c *Control) Render(ctx ui.RenderContext) {
}
