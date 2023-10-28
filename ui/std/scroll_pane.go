package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

// ScrollPaneData holds the available configuration options for the
// ScrollPane component.
type ScrollPaneData struct {

	// DisableHorizontal stops the pane from scrolling horizontally.
	DisableHorizontal bool

	// DisableVertical stops the pane from scrolling vertically.
	DisableVertical bool

	// Focused specifies whether this scroll pane should automatically get
	// the focus.
	Focused bool
}

var scrollPaneDefaultData = ScrollPaneData{}

// ScrollPane is a container component that provides scrolling functionality
// in order to accommodate all children.
var ScrollPane = co.Define(&scrollPaneComponent{})

type scrollPaneComponent struct {
	co.BaseComponent

	canScrollHorizontally bool
	canScrollVertically   bool
	isFocused             bool

	offsetX    float32
	offsetY    float32
	maxOffsetX float32
	maxOffsetY float32
}

func (c *scrollPaneComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), scrollPaneDefaultData)
	c.canScrollHorizontally = !data.DisableHorizontal
	c.canScrollVertically = !data.DisableVertical
	c.isFocused = data.Focused
}

func (c *scrollPaneComponent) Apply(element *ui.Element) {
	var maxChildSize ui.Size

	contentBounds := element.ContentBounds()
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := layout.ElementData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if !c.canScrollHorizontally {
			if layoutConfig.GrowHorizontally {
				childSize.Width = contentBounds.Width
			} else {
				childSize.Width = min(contentBounds.Width, childSize.Width)
			}
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}
		if !c.canScrollVertically {
			if layoutConfig.GrowVertically {
				childSize.Height = contentBounds.Height
			} else {
				childSize.Height = min(contentBounds.Height, childSize.Height)
			}
		}

		maxChildSize = ui.Size{
			Width:  max(maxChildSize.Width, childSize.Width),
			Height: max(maxChildSize.Height, childSize.Height),
		}

		childElement.SetBounds(ui.Bounds{
			Position: ui.NewPosition(-int(c.offsetX), -int(c.offsetY)),
			Size:     childSize,
		})
	}

	c.maxOffsetX = float32(max(0, maxChildSize.Width-contentBounds.Width))
	c.maxOffsetY = float32(max(0, maxChildSize.Height-contentBounds.Height))
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
	c.offsetY = min(max(c.offsetY, 0), c.maxOffsetY)

	element.SetIdealSize(maxChildSize.Grow(element.Padding().Size()))
}

func (c *scrollPaneComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Code {
	case ui.KeyCodeArrowDown:
		if event.Action == ui.KeyboardActionDown || event.Action == ui.KeyboardActionRepeat {
			c.scroll(element, 0.0, -10.0)
			return true
		}
	case ui.KeyCodePageDown:
		if event.Action == ui.KeyboardActionDown || event.Action == ui.KeyboardActionRepeat {
			c.scroll(element, 0.0, -100.0)
			return true
		}
	case ui.KeyCodeArrowUp:
		if event.Action == ui.KeyboardActionDown || event.Action == ui.KeyboardActionRepeat {
			c.scroll(element, 0.0, 10.0)
			return true
		}
	case ui.KeyCodePageUp:
		if event.Action == ui.KeyboardActionDown || event.Action == ui.KeyboardActionRepeat {
			c.scroll(element, 0.0, 100.0)
			return true
		}
	}
	return false
}

func (c *scrollPaneComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	// TODO: Support mouse dragging as a means to scroll
	if event.Action == ui.MouseActionScroll {
		c.scroll(element, event.ScrollX, event.ScrollY)
		return true
	}
	return false
}

func (c *scrollPaneComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithData(co.ElementData{
			Focusable: opt.V(true),
			Focused:   opt.V(c.isFocused),
			Essence:   c,
			Layout:    c,
		})
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithChildren(c.Properties().Children())
	})
}

func (c *scrollPaneComponent) scroll(element *ui.Element, deltaX, deltaY float32) {
	c.offsetX -= deltaX
	c.offsetY -= deltaY
	if c.canScrollHorizontally && !c.canScrollVertically {
		c.offsetX -= deltaY
	}
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
	c.offsetY = min(max(c.offsetY, 0), c.maxOffsetY)

	c.Apply(element)
	element.Invalidate()
}
