package template

import "github.com/mokiat/lacking/ui"

// Initialize wires the template package to the specified Window.
// The specified instance will be the root component used.
func Initialize(window *ui.Window, instance Instance) {
	uiCtx = window.Context()

	rootNode := createComponentNode(New(Element, func() {
		WithData(ElementData{
			Layout: &fillLayout{},
		})
		WithChild("root", instance)
	}))
	window.SetRoot(rootNode.element)
}

type fillLayout struct{}

func (l *fillLayout) Apply(element *ui.Element) {
	for child := element.FirstChild(); child != nil; child = child.RightSibling() {
		child.SetBounds(element.Bounds())
	}
}
