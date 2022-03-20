package mat

import "github.com/mokiat/lacking/ui"

// NewFillLayout creates a new FillLayout instance.
func NewFillLayout() *FillLayout {
	return &FillLayout{}
}

var _ ui.Layout = (*FillLayout)(nil)

// FillLayout is an implementation of Layout that positions and
// resizes elements so that they take up the whole content area.
type FillLayout struct{}

// Apply applies this layout to the specified Element.
func (l *FillLayout) Apply(element *ui.Element) {
	contentBounds := element.ContentBounds()
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		childElement.SetBounds(contentBounds)
	}
	element.SetIdealSize(l.calculateIdealSize(element))
}

func (l *FillLayout) calculateIdealSize(element *ui.Element) ui.Size {
	result := ui.NewSize(0, 0)
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

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
	result.Width += element.Padding().Horizontal()
	result.Height += element.Padding().Vertical()
	return result
}
