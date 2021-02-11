package standard

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.Register("AnchorLayout", ui.BuilderFunc(func(ctx ui.BuildContext) (ui.Control, error) {
		return BuildAnchorLayout(ctx)
	}))
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

func BuildAnchorLayout(ctx ui.BuildContext) (AnchorLayout, error) {
	element := ui.CreateElement()
	layout := &anchorLayout{
		Control: ui.CreateControl(
			ctx.Template.ID,
			element,
			ctx.Template.LayoutAttributes,
		),
	}
	element.SetControl(layout)
	element.SetHandler(layout)

	for _, childTemplate := range ctx.Template.Children {
		child, err := ui.Build(ui.BuildContext{
			Window:   ctx.Window,
			Template: childTemplate,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build child: %w", err)
		}
		layout.AddControl(child)
	}

	return layout, nil
}

type anchorLayout struct {
	ui.Control
}

func (l *anchorLayout) AddControl(control ui.Control) {
	l.Element().Append(control.Element())
	// TODO: Layout child
}

func (l *anchorLayout) RemoveControl(control ui.Control) {
	l.Element().Remove(control.Element())
}

func (l *anchorLayout) OnResize(element *ui.Element, bounds ui.Bounds) {
	for childElement := l.Element().FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		childBounds := ui.Bounds{}
		layoutData := buildAnchorLayoutData(childElement.Control().LayoutAttributes())

		// horizontal
		if layoutData.Width != nil {
			childBounds.Width = *layoutData.Width
			switch {
			case layoutData.Left != nil:
				childBounds.X = l.horizontalPosition(*layoutData.Left, layoutData.LeftRelative)
			case layoutData.Right != nil:
				childBounds.X = l.horizontalPosition(*layoutData.Right, layoutData.RightRelative) - childBounds.Width
			case layoutData.HorizontalCenter != nil:
				childBounds.X = l.horizontalPosition(*layoutData.HorizontalCenter, layoutData.HorizontalCenterRelative) - childBounds.Width/2
			}
		}

		// vertical
		if layoutData.Height != nil {
			childBounds.Height = *layoutData.Height
			switch {
			case layoutData.Top != nil:
				childBounds.Y = l.verticalPosition(*layoutData.Top, layoutData.TopRelative)
			case layoutData.Bottom != nil:
				childBounds.Y = l.verticalPosition(*layoutData.Bottom, layoutData.BottomRelative) - childBounds.Height
			case layoutData.VerticalCenter != nil:
				childBounds.Y = l.verticalPosition(*layoutData.VerticalCenter, layoutData.VerticalCenterRelative) - childBounds.Height/2
			}
		}

		childElement.SetBounds(childBounds)
	}
}

func (l *anchorLayout) OnRender(element *ui.Element, ctx ui.RenderContext) {
	ctx.Canvas.UseSolidColor(ui.RGB(
		byte(rand.Int()%256),
		byte(rand.Int()%256),
		byte(rand.Int()%256),
	))
	ctx.Canvas.DrawRectangle(
		ui.NewPosition(0, 0),
		l.Element().Bounds().Size,
	)
}

func (l *anchorLayout) horizontalPosition(value int, relativeTo anchorRelative) int {
	bounds := l.Element().Bounds()
	switch relativeTo {
	case anchorRelativeLeft:
		return value
	case anchorRelativeRight:
		return value + bounds.Width
	case anchorRelativeCenter:
		return value + bounds.Width/2
	default:
		panic(fmt.Errorf("unexpected horizontal relative to: %d", relativeTo))
	}
}

func (l *anchorLayout) verticalPosition(value int, relativeTo anchorRelative) int {
	bounds := l.Element().Bounds()
	switch relativeTo {
	case anchorRelativeTop:
		return value
	case anchorRelativeBottom:
		return value + bounds.Height
	case anchorRelativeCenter:
		return value + bounds.Height/2
	default:
		panic(fmt.Errorf("unexpected vertical relative to: %d", relativeTo))
	}
}

func buildAnchorLayoutData(attributes ui.AttributeSet) anchorLayoutData {
	var result anchorLayoutData
	if left, ok := attributes.IntAttribute("left"); ok {
		result.Left = &left
	}
	if leftRelative, ok := AnchorRelativeAttribute(attributes, "left-relative"); ok {
		result.LeftRelative = leftRelative
	} else {
		result.LeftRelative = anchorRelativeLeft
	}
	if right, ok := attributes.IntAttribute("right"); ok {
		result.Right = &right
	}
	if rightRelative, ok := AnchorRelativeAttribute(attributes, "right-relative"); ok {
		result.RightRelative = rightRelative
	} else {
		result.RightRelative = anchorRelativeLeft
	}
	if top, ok := attributes.IntAttribute("top"); ok {
		result.Top = &top
	}
	if topRelative, ok := AnchorRelativeAttribute(attributes, "top-relative"); ok {
		result.TopRelative = topRelative
	} else {
		result.TopRelative = anchorRelativeTop
	}
	if bottom, ok := attributes.IntAttribute("bottom"); ok {
		result.Bottom = &bottom
	}
	if bottomRelative, ok := AnchorRelativeAttribute(attributes, "bottom-relative"); ok {
		result.BottomRelative = bottomRelative
	} else {
		result.BottomRelative = anchorRelativeTop
	}
	if horizontalCenter, ok := attributes.IntAttribute("horizontal-center"); ok {
		result.HorizontalCenter = &horizontalCenter
	}
	if horizontalCenterRelative, ok := AnchorRelativeAttribute(attributes, "horizontal-center-relative"); ok {
		result.HorizontalCenterRelative = horizontalCenterRelative
	} else {
		result.HorizontalCenterRelative = anchorRelativeCenter
	}
	if verticalCenter, ok := attributes.IntAttribute("vertical-center"); ok {
		result.VerticalCenter = &verticalCenter
	}
	if verticalCenterRelative, ok := AnchorRelativeAttribute(attributes, "vertical-center-relative"); ok {
		result.VerticalCenterRelative = verticalCenterRelative
	} else {
		result.VerticalCenterRelative = anchorRelativeCenter
	}
	if width, ok := attributes.IntAttribute("width"); ok {
		result.Width = &width
	}
	if height, ok := attributes.IntAttribute("height"); ok {
		result.Height = &height
	}
	return result
}

type anchorLayoutData struct {
	Left                     *int
	LeftRelative             anchorRelative
	Right                    *int
	RightRelative            anchorRelative
	Top                      *int
	TopRelative              anchorRelative
	Bottom                   *int
	BottomRelative           anchorRelative
	HorizontalCenter         *int
	HorizontalCenterRelative anchorRelative
	VerticalCenter           *int
	VerticalCenterRelative   anchorRelative
	Width                    *int
	Height                   *int
}

type anchorRelative int

const (
	anchorRelativeLeft anchorRelative = 1 + iota
	anchorRelativeRight
	anchorRelativeTop
	anchorRelativeBottom
	anchorRelativeCenter
)

func AnchorRelativeAttribute(attributes ui.AttributeSet, name string) (anchorRelative, bool) {
	if value, ok := attributes.StringAttribute(name); ok {
		switch strings.ToLower(value) {
		case "left":
			return anchorRelativeLeft, true
		case "right":
			return anchorRelativeRight, true
		case "top":
			return anchorRelativeTop, true
		case "bottom":
			return anchorRelativeBottom, true
		case "center", "centre":
			return anchorRelativeCenter, true
		}
	}
	return 0, false
}
