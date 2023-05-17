package layout

import "github.com/mokiat/lacking/ui"

// Fill returns a layout that positions elements within the whole bounds of
// the receiver element.
//
// If the receiver element has multiple children then each one is positioned
// independently to take up the whole space.
func Fill() ui.Layout {
	return &fillLayout{}
}

type fillLayout struct{}

func (l *fillLayout) Apply(element *ui.Element) {
	contentBounds := element.ContentBounds()
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		childElement.SetBounds(contentBounds)
	}
	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *fillLayout) calculateIdealSize(element *ui.Element) ui.Size {
	result := ui.NewSize(0, 0)
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}

		result.Width = maxInt(result.Width, childSize.Width)
		result.Height = maxInt(result.Height, childSize.Height)
	}
	return result.Grow(element.Padding().Size())
}
