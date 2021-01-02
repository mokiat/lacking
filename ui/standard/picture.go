package standard

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/behavior"
)

func init() {
	ui.Register("Picture", ui.BuilderFunc(func(ctx ui.BuildContext) (ui.Control, error) {
		return BuildPicture(ctx)
	}))
}

func BuildPicture(ctx ui.BuildContext) (*Picture, error) {
	return &Picture{
		Control: behavior.BuildControl(ctx),
	}, nil
}

type Picture struct {
	*behavior.Control
}
