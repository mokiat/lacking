package container

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/control"
)

type Basic struct {
	control.Basic
	children []ui.Control
}

var _ ui.Container = (*Basic)(nil)

func (b *Basic) OnUpdateLayout(bounds ui.Bounds) {
}

func (b *Basic) OnRender(canvas ui.Canvas, dirtyBounds ui.Bounds) {
	for _, child := range b.children {
		canvas.Push()
		child.OnRender(canvas, child.Bounds())
		canvas.Pop()
	}
}
