package ui

type Control interface {
	SetBounds(bounds Bounds)
	Bounds() Bounds
	OnRender(canvas Canvas, dirtyBounds Bounds)
}
