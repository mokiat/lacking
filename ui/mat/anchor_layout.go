package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
)

// AnchorLayoutData represents a layout configuration for a component
// that is added to a Container with layout set to AnchorLayout.
type AnchorLayoutData struct {
	Left             optional.Int
	Right            optional.Int
	Top              optional.Int
	Bottom           optional.Int
	HorizontalCenter optional.Int
	VerticalCenter   optional.Int
	Width            optional.Int
	Height           optional.Int
}

// AnchorLayoutSettings contains optional configurations for the
// AnchorLayout.
type AnchorLayoutSettings struct{}

// NewAnchorLayout creates a new AnchorLayout instance.
func NewAnchorLayout(settings AnchorLayoutSettings) *AnchorLayout {
	return &AnchorLayout{}
}

var _ ui.Layout = (*AnchorLayout)(nil)

// AnchorLayout is an implementation of Layout that positions and
// resizes elements according to border-relative coordinates.
type AnchorLayout struct{}

// Apply applies this layout to the specified Element.
func (l *AnchorLayout) Apply(element *ui.Element) {
	// TODO: Consider children's margin settings

	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := childElement.LayoutConfig().(AnchorLayoutData)
		childBounds := ui.Bounds{}

		// horizontal
		if layoutConfig.Width.Specified {
			childBounds.Width = layoutConfig.Width.Value
			switch {
			case layoutConfig.Left.Specified:
				childBounds.X = l.leftPosition(element, layoutConfig.Left.Value)
			case layoutConfig.Right.Specified:
				childBounds.X = l.rightPosition(element, layoutConfig.Right.Value) - childBounds.Width
			case layoutConfig.HorizontalCenter.Specified:
				childBounds.X = l.horizontalCenterPosition(element, layoutConfig.HorizontalCenter.Value) - childBounds.Width/2
			}
		}

		// vertical
		if layoutConfig.Height.Specified {
			childBounds.Height = layoutConfig.Height.Value
			switch {
			case layoutConfig.Top.Specified:
				childBounds.Y = l.topPosition(element, layoutConfig.Top.Value)
			case layoutConfig.Bottom.Specified:
				childBounds.Y = l.bottomPosition(element, layoutConfig.Bottom.Value) - childBounds.Height
			case layoutConfig.VerticalCenter.Specified:
				childBounds.Y = l.verticalCenterPosition(element, layoutConfig.VerticalCenter.Value) - childBounds.Height/2
			}
		}

		if layoutConfig.Left.Specified {
			childBounds.X = l.leftPosition(element, layoutConfig.Left.Value)
			switch {
			case layoutConfig.Right.Specified:
				childBounds.Width = l.rightPosition(element, layoutConfig.Right.Value) - childBounds.X
			}
		}

		if layoutConfig.Top.Specified {
			childBounds.Y = l.topPosition(element, layoutConfig.Top.Value)
			switch {
			case layoutConfig.Bottom.Specified:
				childBounds.Height = l.bottomPosition(element, layoutConfig.Bottom.Value) - childBounds.Y
			}
		}

		childElement.SetBounds(childBounds)
	}
}

func (l *AnchorLayout) leftPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.X + value
}

func (l *AnchorLayout) rightPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.X + bounds.Width - value
}

func (l *AnchorLayout) topPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.Y + value
}

func (l *AnchorLayout) bottomPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.Y + bounds.Height - value
}

func (l *AnchorLayout) horizontalCenterPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.X + bounds.Width/2 + value
}

func (l *AnchorLayout) verticalCenterPosition(element *ui.Element, value int) int {
	bounds := element.ContentBounds()
	return bounds.Y + bounds.Height/2 + value
}
