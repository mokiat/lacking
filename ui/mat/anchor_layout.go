package mat

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

// AnchorLayoutData represents a layout configuration for a component
// that is added to a Container with layout set to AnchorLayout.
type AnchorLayoutData struct {
	Left                     *int
	LeftRelation             Relation
	Right                    *int
	RightRelation            Relation
	Top                      *int
	TopRelation              Relation
	Bottom                   *int
	BottomRelation           Relation
	HorizontalCenter         *int
	HorizontalCenterRelation Relation
	VerticalCenter           *int
	VerticalCenterRelation   Relation
	Width                    *int
	Height                   *int
}

// AnchorLayoutSettings contains optional configurations for the
// AnchorLayout.
type AnchorLayoutSettings struct{}

// NewAnchorLayout creates a new AnchorLayout instance.
func NewAnchorLayout(settings AnchorLayoutSettings) *AnchorLayout {
	return &AnchorLayout{}
}

var _ Layout = (*AnchorLayout)(nil)

// AnchorLayout is an implementation of Layout that positions and
// resizes elements according to border-relative coordinates.
type AnchorLayout struct{}

// Apply applies this layout to the specified Element.
func (l *AnchorLayout) Apply(element *ui.Element) {
	// TODO: Consider content area
	// TODO: Consider children's margin settings

	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := childElement.LayoutConfig().(AnchorLayoutData)
		childBounds := ui.Bounds{}

		// horizontal
		if layoutConfig.Width != nil {
			childBounds.Width = *layoutConfig.Width
			switch {
			case layoutConfig.Left != nil:
				childBounds.X = l.horizontalPosition(element, *layoutConfig.Left, layoutConfig.LeftRelation)
			case layoutConfig.Right != nil:
				childBounds.X = l.horizontalPosition(element, *layoutConfig.Right, layoutConfig.RightRelation) - childBounds.Width
			case layoutConfig.HorizontalCenter != nil:
				childBounds.X = l.horizontalPosition(element, *layoutConfig.HorizontalCenter, layoutConfig.HorizontalCenterRelation) - childBounds.Width/2
			}
		}

		// vertical
		if layoutConfig.Height != nil {
			childBounds.Height = *layoutConfig.Height
			switch {
			case layoutConfig.Top != nil:
				childBounds.Y = l.verticalPosition(element, *layoutConfig.Top, layoutConfig.TopRelation)
			case layoutConfig.Bottom != nil:
				childBounds.Y = l.verticalPosition(element, *layoutConfig.Bottom, layoutConfig.BottomRelation) - childBounds.Height
			case layoutConfig.VerticalCenter != nil:
				childBounds.Y = l.verticalPosition(element, *layoutConfig.VerticalCenter, layoutConfig.VerticalCenterRelation) - childBounds.Height/2
			}
		}

		if layoutConfig.Left != nil {
			childBounds.X = l.horizontalPosition(element, *layoutConfig.Left, layoutConfig.LeftRelation)
			switch {
			case layoutConfig.Right != nil:
				childBounds.Width = l.horizontalPosition(element, *layoutConfig.Right, layoutConfig.RightRelation) - childBounds.X
			}
		}

		if layoutConfig.Top != nil {
			childBounds.Y = l.verticalPosition(element, *layoutConfig.Top, layoutConfig.TopRelation)
			switch {
			case layoutConfig.Bottom != nil:
				childBounds.Height = l.verticalPosition(element, *layoutConfig.Bottom, layoutConfig.BottomRelation) - childBounds.Y
			}
		}

		childElement.SetBounds(childBounds)
	}
}

func (l *AnchorLayout) horizontalPosition(element *ui.Element, value int, relativeTo Relation) int {
	bounds := element.Bounds()
	switch relativeTo {
	case RelationLeft:
		return value
	case RelationRight:
		return value + bounds.Width
	case RelationCenter:
		return value + bounds.Width/2
	default:
		panic(fmt.Errorf("unexpected horizontal relative to: %d", relativeTo))
	}
}

func (l *AnchorLayout) verticalPosition(element *ui.Element, value int, relativeTo Relation) int {
	bounds := element.Bounds()
	switch relativeTo {
	case RelationTop:
		return value
	case RelationBottom:
		return value + bounds.Height
	case RelationCenter:
		return value + bounds.Height/2
	default:
		panic(fmt.Errorf("unexpected vertical relative to: %d", relativeTo))
	}
}
