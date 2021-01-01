package standard

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/behavior"
)

func init() {
	ui.Register("Picture", ui.BuilderFunc(func(ctx ui.BuildContext) ui.Control {
		return BuildPicture(ctx)
	}))
}

func BuildPicture(ctx ui.BuildContext) *Picture {
	return &Picture{
		Control: behavior.BuildControl(ctx),
	}
}

type Picture struct {
	*behavior.Control
}
