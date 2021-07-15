package template

import (
	"github.com/mokiat/lacking/ui"
)

var uiCtx *ui.Context

// Initialize wires the template package to the specified view.
// The specified init function is used to construct the root
// component.
func Initialize(view *ui.View, instance Instance) {
	app := &application{
		instance:  instance,
		hierarchy: &hierarchy{},
	}
	view.SetHandler(app)
}

var _ ui.ElementResizeHandler = (*application)(nil)

type application struct {
	instance  Instance
	hierarchy *hierarchy
	rootNode  *componentNode
}

func (a *application) OnCreate(view *ui.View) {
	uiCtx = view.Context()

	a.rootNode = a.hierarchy.CreateComponentNode(New(Element, func() {
		WithData(ElementData{
			Essence: a,
		})
		WithChild("root", a.instance)
	}))
	view.SetRoot(a.rootNode.Element())
}

func (a *application) OnShow(view *ui.View) {}

func (a *application) OnHide(view *ui.View) {}

func (a *application) OnDestroy(view *ui.View) {
	view.SetRoot(nil)
	a.hierarchy.DestroyComponentNode(a.rootNode)
	a.rootNode = nil

	uiCtx = nil
}

func (a *application) OnResize(element *ui.Element, bounds ui.Bounds) {
	for child := element.FirstChild(); child != nil; child = child.RightSibling() {
		child.SetBounds(bounds)
	}
}
