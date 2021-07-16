package mat

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
)

// AnchorLayoutData represents a layout configuration for a component
// that is added to a Container with layout set to AnchorLayout.
type AnchorLayoutData struct {
	Left                     optional.Int
	LeftRelation             Relation
	Right                    optional.Int
	RightRelation            Relation
	Top                      optional.Int
	TopRelation              Relation
	Bottom                   optional.Int
	BottomRelation           Relation
	HorizontalCenter         optional.Int
	HorizontalCenterRelation Relation
	VerticalCenter           optional.Int
	VerticalCenterRelation   Relation
	Width                    optional.Int
	Height                   optional.Int
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
		if layoutConfig.Width.Specified {
			childBounds.Width = layoutConfig.Width.Value
			switch {
			case layoutConfig.Left.Specified:
				childBounds.X = l.horizontalPosition(element, layoutConfig.Left.Value, layoutConfig.LeftRelation)
			case layoutConfig.Right.Specified:
				childBounds.X = l.horizontalPosition(element, layoutConfig.Right.Value, layoutConfig.RightRelation) - childBounds.Width
			case layoutConfig.HorizontalCenter.Specified:
				childBounds.X = l.horizontalPosition(element, layoutConfig.HorizontalCenter.Value, layoutConfig.HorizontalCenterRelation) - childBounds.Width/2
			}
		}

		// vertical
		if layoutConfig.Height.Specified {
			childBounds.Height = layoutConfig.Height.Value
			switch {
			case layoutConfig.Top.Specified:
				childBounds.Y = l.verticalPosition(element, layoutConfig.Top.Value, layoutConfig.TopRelation)
			case layoutConfig.Bottom.Specified:
				childBounds.Y = l.verticalPosition(element, layoutConfig.Bottom.Value, layoutConfig.BottomRelation) - childBounds.Height
			case layoutConfig.VerticalCenter.Specified:
				childBounds.Y = l.verticalPosition(element, layoutConfig.VerticalCenter.Value, layoutConfig.VerticalCenterRelation) - childBounds.Height/2
			}
		}

		if layoutConfig.Left.Specified {
			childBounds.X = l.horizontalPosition(element, layoutConfig.Left.Value, layoutConfig.LeftRelation)
			switch {
			case layoutConfig.Right.Specified:
				childBounds.Width = l.horizontalPosition(element, layoutConfig.Right.Value, layoutConfig.RightRelation) - childBounds.X
			}
		}

		if layoutConfig.Top.Specified {
			childBounds.Y = l.verticalPosition(element, layoutConfig.Top.Value, layoutConfig.TopRelation)
			switch {
			case layoutConfig.Bottom.Specified:
				childBounds.Height = l.verticalPosition(element, layoutConfig.Bottom.Value, layoutConfig.BottomRelation) - childBounds.Y
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
