package template

import (
	"github.com/mokiat/lacking/ui"
)

type InitContext struct {
	*Context
}

type InitFunc func(ctx InitContext) Instance

// Initialize wires the template package to the specified view.
// The specified init function is used to construct the root
// component.
func Initialize(view *ui.View, initFn InitFunc) {
	hierarchy := &hierarchy{
		uiCtx: view.Context(),
	}
	app := &application{
		initFn:    initFn,
		hierarchy: hierarchy,
	}
	view.SetHandler(app)
}

var _ ui.ElementResizeHandler = (*application)(nil)

type application struct {
	initFn    InitFunc
	hierarchy *hierarchy
	rootNode  *componentNode
}

func (a *application) OnCreate(view *ui.View) {
	ctx := &Context{}
	a.rootNode = a.hierarchy.CreateComponentNode(ctx.Instance(Element, "root", func() {
		ctx.WithData(ElementData{
			Essence: a,
		})
		ctx.WithChild(a.initFn(InitContext{
			Context: ctx,
		}))
	}))
	view.SetRoot(a.rootNode.Element())
}

func (a *application) OnShow(view *ui.View) {}

func (a *application) OnHide(view *ui.View) {}

func (a *application) OnDestroy(view *ui.View) {
	view.SetRoot(nil)
	a.hierarchy.DestroyComponentNode(a.rootNode)
	a.rootNode = nil
}

func (a *application) OnResize(element *ui.Element, bounds ui.Bounds) {
	for child := element.FirstChild(); child != nil; child = child.RightSibling() {
		child.SetBounds(bounds)
	}
}
