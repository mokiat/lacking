package ui

type Control interface {
	ID() string
	SetBounds(bounds Bounds)
	Bounds() Bounds
	Render(ctx RenderContext)
}
