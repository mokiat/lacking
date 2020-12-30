package ui

type Container interface {
	Control
	OnUpdateLayout(bounds Bounds)
}
