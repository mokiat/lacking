package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("AnchorLayout", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildAnchorLayout(ctx, template, layoutConfig)
	}))
}

// AnchorLayoutConfig represents a layout configuration for a control
// that is added to an AnchorLayout.
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
		c.RightRelation = RelationLeft
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
		c.BottomRelation = RelationTop
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

// AnchorLayout represents a layout that positions
// controls based on the following anchor references:
//     left - the left side of the layout
//     right - the right side of the layout
//     top - the top side of the layout
//     bottom - the bottom side of the layout
//     center - the center of the layout
type AnchorLayout interface {
	ui.Control

	// AddControl adds a control to this layout.
	AddControl(control ui.Control)

	// RemoveControl removes the specified control from this layout.
	RemoveControl(control ui.Control)
}

func BuildAnchorLayout(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (AnchorLayout, error) {
	layout := &anchorLayout{}

	element := ctx.CreateElement()
	element.SetLayoutConfig(layoutConfig)
	element.SetHandler(layout)

	layout.Control = ctx.CreateControl(element)
	element.SetControl(layout)
	if err := layout.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}

	for _, childTemplate := range template.Children() {
		childLayoutConfig := new(AnchorLayoutConfig)
		childLayoutConfig.ApplyAttributes(childTemplate.LayoutAttributes())
		child, err := ctx.InstantiateTemplate(childTemplate, childLayoutConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate child from template: %w", err)
		}
		layout.AddControl(child)
	}

	return layout, nil
}

var _ ui.ElementResizeHandler = (*anchorLayout)(nil)
var _ ui.ElementRenderHandler = (*anchorLayout)(nil)

type anchorLayout struct {
	ui.Control
	backgroundColor *ui.Color
}

func (l *anchorLayout) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := l.Element().ApplyAttributes(attributes); err != nil {
		return err
	}
	if color, ok := attributes.ColorAttribute("background-color"); ok {
		l.backgroundColor = &color
	}
	return nil
}

func (l *anchorLayout) AddControl(control ui.Control) {
	l.Element().AppendChild(control.Element())
}

func (l *anchorLayout) RemoveControl(control ui.Control) {
	l.Element().RemoveChild(control.Element())
}

func (l *anchorLayout) OnResize(element *ui.Element, bounds ui.Bounds) {
	// TODO: Consider content area
	// TODO: Consider children's margin settings

	for childElement := l.Element().FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := childElement.LayoutConfig().(*AnchorLayoutConfig)
		childBounds := ui.Bounds{}

		// horizontal
		if layoutConfig.Width != nil {
			childBounds.Width = *layoutConfig.Width
			switch {
			case layoutConfig.Left != nil:
				childBounds.X = l.horizontalPosition(*layoutConfig.Left, layoutConfig.LeftRelation)
			case layoutConfig.Right != nil:
				childBounds.X = l.horizontalPosition(*layoutConfig.Right, layoutConfig.RightRelation) - childBounds.Width
			case layoutConfig.HorizontalCenter != nil:
				childBounds.X = l.horizontalPosition(*layoutConfig.HorizontalCenter, layoutConfig.HorizontalCenterRelation) - childBounds.Width/2
			}
		}

		// vertical
		if layoutConfig.Height != nil {
			childBounds.Height = *layoutConfig.Height
			switch {
			case layoutConfig.Top != nil:
				childBounds.Y = l.verticalPosition(*layoutConfig.Top, layoutConfig.TopRelation)
			case layoutConfig.Bottom != nil:
				childBounds.Y = l.verticalPosition(*layoutConfig.Bottom, layoutConfig.BottomRelation) - childBounds.Height
			case layoutConfig.VerticalCenter != nil:
				childBounds.Y = l.verticalPosition(*layoutConfig.VerticalCenter, layoutConfig.VerticalCenterRelation) - childBounds.Height/2
			}
		}

		if layoutConfig.Left != nil {
			childBounds.X = l.horizontalPosition(*layoutConfig.Left, layoutConfig.LeftRelation)
			switch {
			case layoutConfig.Right != nil:
				childBounds.Width = l.horizontalPosition(*layoutConfig.Right, layoutConfig.RightRelation) - childBounds.X
			}
		}

		if layoutConfig.Top != nil {
			childBounds.Y = l.verticalPosition(*layoutConfig.Top, layoutConfig.TopRelation)
			switch {
			case layoutConfig.Bottom != nil:
				childBounds.Height = l.verticalPosition(*layoutConfig.Bottom, layoutConfig.BottomRelation) - childBounds.Y
			}
		}

		childElement.SetBounds(childBounds)
	}
}

func (l *anchorLayout) OnRender(element *ui.Element, canvas ui.Canvas) {
	if l.backgroundColor != nil {
		canvas.SetSolidColor(*l.backgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			l.Element().Bounds().Size,
		)
	}
}

func (l *anchorLayout) horizontalPosition(value int, relativeTo Relation) int {
	bounds := l.Element().Bounds()
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

func (l *anchorLayout) verticalPosition(value int, relativeTo Relation) int {
	bounds := l.Element().Bounds()
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
