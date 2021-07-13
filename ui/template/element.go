package template

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
)

const namespace = "github.com/mokiat/lacking/ui/template"

type ElementData struct {
	Essence      ui.Essence
	Enabled      optional.Bool
	Visible      optional.Bool
	Materialized optional.Bool
}

var Element = NewComponentType(namespace, "Element", newElement)

func newElement(ctx *ui.Context) Component {
	return &element{
		ctx: ctx,
	}
}

type element struct {
	NopComponent
	ctx     *ui.Context
	element *ui.Element
}

func (e *element) OnCreated() {
	e.element = e.ctx.CreateElement()
}

func (e *element) OnDestroyed() {
	e.element.Destroy()
}

func (e *element) Render(ctx RenderContext) Instance {
	data := ctx.Data().(ElementData)
	if data.Essence != nil {
		e.element.SetEssence(data.Essence)
	}
	if data.Enabled.Specified {
		e.element.SetEnabled(data.Enabled.Value)
	}
	if data.Visible.Specified {
		e.element.SetVisible(data.Visible.Value)
	}
	if data.Materialized.Specified {
		e.element.SetMaterialized(data.Materialized.Value)
	}

	e.element.SetLayoutConfig(ctx.LayoutData())

	return Instance{
		element:  e.element,
		children: ctx.Children(),
	}
}
