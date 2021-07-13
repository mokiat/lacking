package template

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

type ComponentConstructor func(ctx *ui.Context) Component

type Component interface {
	// OnCreated is called immediatelly after the component instance
	// has been created but prior to it being mounted.
	OnCreated()

	// OnDestroyed is called after the component instance has bee
	// unmounted and at the point in which the framework will give
	// up on it.
	OnDestroyed()

	OnDataChanged(data interface{})

	OnLayoutDataChanged(layoutData interface{})

	OnCallbackDataChanged(callbackData interface{})

	OnChildrenChanged(children []Instance)

	Render(ctx RenderContext) Instance
}

type ComponentType interface {
	Name() string
	NewComponent(ctx *ui.Context) Component
}

func NewComponentType(namespace, name string, constructor ComponentConstructor) ComponentType {
	return &internalComponentType{
		name:        fmt.Sprintf("%s.%s", namespace, name),
		constructor: constructor,
	}
}

type internalComponentType struct {
	name        string
	constructor ComponentConstructor
}

func (t *internalComponentType) Name() string {
	return t.name
}

func (t *internalComponentType) NewComponent(ctx *ui.Context) Component {
	return t.constructor(ctx)
}

var _ Component = NopComponent{}

type NopComponent struct{}

func (c NopComponent) OnCreated() {}

func (c NopComponent) OnDestroyed() {}

func (c NopComponent) OnDataChanged(data interface{}) {}

func (c NopComponent) OnLayoutDataChanged(layoutData interface{}) {}

func (c NopComponent) OnCallbackDataChanged(callbackData interface{}) {}

func (c NopComponent) OnChildrenChanged(children []Instance) {}

func (c NopComponent) Render(ctx RenderContext) Instance {
	panic("rendering not implemented for this component")
}

type FunctionalComponentFunc func(ctx RenderContext) Instance

func FunctionalComponent(fn FunctionalComponentFunc) Component {
	return &internalFunctionalComponent{
		fn: fn,
	}
}

type internalFunctionalComponent struct {
	NopComponent
	fn FunctionalComponentFunc
}

func (c internalFunctionalComponent) Render(ctx RenderContext) Instance {
	return c.fn(ctx)
}
