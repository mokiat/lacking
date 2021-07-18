package template

import (
	"github.com/mokiat/lacking/ui"
)

var uiCtx *ui.Context

// Initialize wires the template package to the specified Window.
// The specified instance will be the root component used.
func Initialize(window *ui.Window, instance Instance) {
	uiCtx = window.Context()

	rootNode := createComponentNode(New(Element, func() {
		WithData(ElementData{
			Essence: &rootElementEssence{},
		})
		WithChild("root", instance)
	}))
	window.SetRoot(rootNode.element)
}

var _ ui.ElementResizeHandler = (*rootElementEssence)(nil)

type rootElementEssence struct{}

func (*rootElementEssence) OnResize(element *ui.Element, bounds ui.Bounds) {
	for child := element.FirstChild(); child != nil; child = child.RightSibling() {
		child.SetBounds(bounds)
	}
}
