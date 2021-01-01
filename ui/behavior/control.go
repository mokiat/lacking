package behavior

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
)

const (
	defaultControlWidth  = 64.0
	defaultControlHeight = 32.0
)

func BuildControl(ctx ui.BuildContext) *Control {
	return &Control{
		id: ctx.Template.ID(),
		bounds: ui.Bounds{
			Position: sprec.ZeroVec2(),
			Size:     sprec.NewVec2(defaultControlWidth, defaultControlHeight),
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
