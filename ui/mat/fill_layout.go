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
		childMargin := childElement.Margin()
		childElement.SetBounds(contentBounds.
			Translate(ui.NewPosition(childMargin.Left, childMargin.Top)).
			Shrink(ui.NewSize(childMargin.Horizontal(), childMargin.Vertical())),
		)
	}
}
