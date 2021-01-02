package standard

import (
	"fmt"
	"strings"

	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/behavior"
)

func init() {
	ui.Register("AnchorLayout", ui.BuilderFunc(func(ctx ui.BuildContext) (ui.Control, error) {
		return BuildAnchorLayout(ctx)
	}))
}

func BuildAnchorLayout(ctx ui.BuildContext) (*AnchorLayout, error) {
	children := make([]ui.Control, len(ctx.Template.Children()))
	for i, childTemplate := range ctx.Template.Children() {
		child, err := ui.Build(ui.BuildContext{
			Template:   childTemplate,
			LayoutData: BuildAnchorLayoutData(childTemplate),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build child: %w", err)
		}
		children[i] = child
	}

	return &AnchorLayout{
		Control:  behavior.BuildControl(ctx),
		children: children,
	}, nil
}

type AnchorLayout struct {
	*behavior.Control
	children []ui.Control
}

func (l *AnchorLayout) SetBounds(bounds ui.Bounds) {
	l.Control.SetBounds(bounds)

	// TODO: Relayout children
}

func BuildAnchorLayoutData(template ui.Template) AnchorLayoutData {
	var result AnchorLayoutData
	if left, ok := template.IntAttribute("layout-left"); ok {
		result.Left = &left
	}
	if leftRelative, ok := AnchorRelativeAttribute(template, "layout-left-relative"); ok {
		result.LeftRelative = leftRelative
	} else {
		result.LeftRelative = AnchorRelativeLeft
	}
	if right, ok := template.IntAttribute("layout-right"); ok {
		result.Right = &right
	}
	if rightRelative, ok := AnchorRelativeAttribute(template, "layout-right-relative"); ok {
		result.RightRelative = rightRelative
	} else {
		result.RightRelative = AnchorRelativeLeft
	}
	if top, ok := template.IntAttribute("layout-top"); ok {
		result.Top = &top
	}
	if topRelative, ok := AnchorRelativeAttribute(template, "layout-top-relative"); ok {
		result.TopRelative = topRelative
	} else {
		result.TopRelative = AnchorRelativeTop
	}
	if bottom, ok := template.IntAttribute("layout-bottom"); ok {
		result.Bottom = &bottom
	}
	if bottomRelative, ok := AnchorRelativeAttribute(template, "layout-bottom-relative"); ok {
		result.BottomRelative = bottomRelative
	} else {
		result.BottomRelative = AnchorRelativeTop
	}
	if width, ok := template.IntAttribute("layout-width"); ok {
		result.Width = &width
	}
	if height, ok := template.IntAttribute("layout-height"); ok {
		result.Height = &height
	}
	return result
}

type AnchorLayoutData struct {
	Left                     *int
	LeftRelative             AnchorRelative
	Right                    *int
	RightRelative            AnchorRelative
	Top                      *int
	TopRelative              AnchorRelative
	Bottom                   *int
	BottomRelative           AnchorRelative
	HorizontalCenter         *int
	HorizontalCenterRelative AnchorRelative
	VerticalCenter           *int
	VerticalCenterRelative   AnchorRelative
	Width                    *int
	Height                   *int
}

type AnchorRelative int

const (
	AnchorRelativeLeft AnchorRelative = 1 + iota
	AnchorRelativeRight
	AnchorRelativeTop
	AnchorRelativeBottom
	AnchorRelativeCenter
)

func AnchorRelativeAttribute(template ui.Template, name string) (AnchorRelative, bool) {
	if value, ok := template.StringAttribute(name); ok {
		switch strings.ToLower(value) {
		case "left":
			return AnchorRelativeLeft, true
		case "right":
			return AnchorRelativeRight, true
		case "top":
			return AnchorRelativeTop, true
		case "bottom":
			return AnchorRelativeBottom, true
		case "center", "centre":
			return AnchorRelativeCenter, true
		}
	}
	return 0, false
}
