package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterLayoutBuilder("AnchorLayout", ui.LayoutBuilderFunc(func(attributes ui.AttributeSet) (ui.Layout, error) {
		return NewAnchorLayout(attributes), nil
	}))
}

// NewAnchorLayout creates a new anchor layout instance.
func NewAnchorLayout(attributes ui.AttributeSet) *AnchorLayout {
	return &AnchorLayout{}
}

var _ ui.Layout = (*AnchorLayout)(nil)

type AnchorLayout struct{}

// LayoutConfig creates a new layout config instance specific
// to this layout.
func (l *AnchorLayout) LayoutConfig() ui.LayoutConfig {
	return &AnchorLayoutConfig{}
}

// Apply applies this layout to the specified Element.
func (l *AnchorLayout) Apply(element *ui.Element) {
	// TODO: Consider content area
	// TODO: Consider children's margin settings

	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := childElement.LayoutConfig().(*AnchorLayoutConfig)
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

// AnchorLayoutConfig represents a layout configuration for a control
// that is added to an Anchor.
type AnchorLayoutConfig struct {
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

func (c *AnchorLayoutConfig) ApplyAttributes(attributes ui.AttributeSet) {
	if left, ok := attributes.IntAttribute("left"); ok {
		c.Left = &left
	}
	if leftRelation, ok := RelationAttribute(attributes, "left-relation"); ok {
		c.LeftRelation = leftRelation
	} else {
		c.LeftRelation = RelationLeft
	}
	if right, ok := attributes.IntAttribute("right"); ok {
		c.Right = &right
	}
	if rightRelation, ok := RelationAttribute(attributes, "right-relation"); ok {
		c.RightRelation = rightRelation
	} else {
		c.RightRelation = RelationRight
	}
	if top, ok := attributes.IntAttribute("top"); ok {
		c.Top = &top
	}
	if topRelation, ok := RelationAttribute(attributes, "top-relation"); ok {
		c.TopRelation = topRelation
	} else {
		c.TopRelation = RelationTop
	}
	if bottom, ok := attributes.IntAttribute("bottom"); ok {
		c.Bottom = &bottom
	}
	if bottomRelation, ok := RelationAttribute(attributes, "bottom-relation"); ok {
		c.BottomRelation = bottomRelation
	} else {
		c.BottomRelation = RelationBottom
	}
	if horizontalCenter, ok := attributes.IntAttribute("horizontal-center"); ok {
		c.HorizontalCenter = &horizontalCenter
	}
	if horizontalCenterRelation, ok := RelationAttribute(attributes, "horizontal-center-relation"); ok {
		c.HorizontalCenterRelation = horizontalCenterRelation
	} else {
		c.HorizontalCenterRelation = RelationCenter
	}
	if verticalCenter, ok := attributes.IntAttribute("vertical-center"); ok {
		c.VerticalCenter = &verticalCenter
	}
	if verticalCenterRelation, ok := RelationAttribute(attributes, "vertical-center-relation"); ok {
		c.VerticalCenterRelation = verticalCenterRelation
	} else {
		c.VerticalCenterRelation = RelationCenter
	}
	if width, ok := attributes.IntAttribute("width"); ok {
		c.Width = &width
	}
	if height, ok := attributes.IntAttribute("height"); ok {
		c.Height = &height
	}
}
