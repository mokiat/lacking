package component

import "github.com/mokiat/lacking/ui"

// Initialize wires the framewprl to the specified ui Window.
// The specified instance will be the root component used.
func Initialize(window *ui.Window, instance Instance) {
	uiCtx = window.Context()

	rootNode := createComponentNode(New(Element, func() {
		WithData(ElementData{
			Layout: ui.NewFillLayout(),
		})
		WithChild("root", instance)
	}))
	window.Root().AppendChild(rootNode.element)
}
