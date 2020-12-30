package control

import "github.com/mokiat/lacking/ui"

type Basic struct {
	bounds ui.Bounds
}

var _ ui.Control = (*Basic)(nil)

func (b *Basic) SetBounds(bounds ui.Bounds) {
	b.bounds = bounds
}

func (b *Basic) Bounds() ui.Bounds {
	return b.bounds
}

func (b *Basic) OnRender(cavans ui.Canvas, dirtyBounds ui.Bounds) {
}
