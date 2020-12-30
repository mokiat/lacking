package container

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/control"
)

type AnchorLayout struct {
	control.Basic
	children []ui.Control
}

var _ ui.Container = (*AnchorLayout)(nil)

func (l *AnchorLayout) OnUpdateLayout(bounds ui.Bounds) {

}
