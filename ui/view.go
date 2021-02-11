package ui

type ViewProvider interface {
	CreateView(window Window) (View, error)
}

type View interface {
	Element() *Element
	// OpenOverlay()
}

func CreateView(child Control) View {
	root := CreateElement()
	if child != nil {
		root.Append(child.Element())
	}
	view := &view{
		root: root,
	}
	root.SetHandler(view)
	return view
}

type view struct {
	root *Element
}

func (v *view) Element() *Element {
	return v.root
}

func (v *view) OnResize(element *Element, bounds Bounds) {
	childElement := v.root.firstChild
	for childElement != nil {
		childElement.SetBounds(bounds)
		childElement = childElement.rightSibling
	}
}
