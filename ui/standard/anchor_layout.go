package standard

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/behavior"
)

func init() {
	ui.Register("AnchorLayout", ui.BuilderFunc(func(ctx ui.BuildContext) ui.Control {
		return BuildAnchorLayout(ctx)
	}))
}

func BuildAnchorLayout(ctx ui.BuildContext) *AnchorLayout {
	return &AnchorLayout{
		Control: behavior.BuildControl(ctx),
	}
}

type AnchorLayout struct {
	*behavior.Control
	children []ui.Control
}

func (l *AnchorLayout) SetBounds(bounds ui.Bounds) {
	l.Control.SetBounds(bounds)

	// TODO: Relayout children
}
