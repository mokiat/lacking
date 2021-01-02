package ui

import "fmt"

var registry map[string]Builder

func init() {
	registry = make(map[string]Builder)
}

func Register(name string, builder Builder) {
	registry[name] = builder
}

type BuildContext struct {
	Template   Template
	LayoutData LayoutData
}

type Builder interface {
	Build(ctx BuildContext) (Control, error)
}

type BuilderFunc func(ctx BuildContext) (Control, error)

func (f BuilderFunc) Build(ctx BuildContext) (Control, error) {
	return f(ctx)
}

func Build(ctx BuildContext) (Control, error) {
	builder, ok := registry[ctx.Template.Name()]
	if !ok {
		return nil, fmt.Errorf("could not find builder for %q", ctx.Template.Name())
	}
	control, err := builder.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build %q: %w", ctx.Template.Name(), err)
	}
	return control, nil
}
