package ui

type BuildContext struct {
	Template   Template
	LayoutData LayoutData
}

type Builder interface {
	Build(ctx BuildContext) Control
}

type BuilderFunc func(ctx BuildContext) Control

func (f BuilderFunc) Build(ctx BuildContext) Control {
	return f(ctx)
}
