package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

// ScrollPaneData holds the available configuration options for the
// ScrollPane component.
type ScrollPaneData struct {

	// DisableHorizontal stops the pane from scrolling horizontally.
	DisableHorizontal bool

	// DisableVertical stops the pane from scrolling vertically.
	DisableVertical bool
}

var defaultScrollPaneData = ScrollPaneData{
	DisableHorizontal: false,
	DisableVertical:   false,
}

var ScrollPane = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data ScrollPaneData
	)
	props.InjectOptionalData(&data, defaultScrollPaneData)

	essence := co.UseState(func() *scrollPaneEssence {
		return &scrollPaneEssence{}
	}).Get()

	essence.scrollHorizontally = !data.DisableHorizontal
	essence.scrollVertically = !data.DisableVertical

	return co.New(Element, func() {
		co.WithData(co.ElementData{
			Focusable: opt.V(true),
			Essence:   essence,
			Layout:    essence,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.Layout = (*scrollPaneEssence)(nil)
var _ ui.ElementKeyboardHandler = (*scrollPaneEssence)(nil)
var _ ui.ElementMouseHandler = (*scrollPaneEssence)(nil)

type scrollPaneEssence struct {
	scrollHorizontally bool
	scrollVertically   bool

	offsetX    float64
	offsetY    float64
	maxOffsetX float64
	maxOffsetY float64
}

func (e *scrollPaneEssence) Apply(element *ui.Element) {
	var maxChildSize ui.Size

	contentBounds := element.ContentBounds()
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if !e.scrollHorizontally && layoutConfig.GrowHorizontally {
			childSize.Width = maxInt(childSize.Width, contentBounds.Width)
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}
		if !e.scrollVertically && layoutConfig.GrowVertically {
			childSize.Height = maxInt(childSize.Height, contentBounds.Height)
		}

		maxChildSize = ui.Size{
			Width:  maxInt(maxChildSize.Width, childSize.Width),
			Height: maxInt(maxChildSize.Height, childSize.Height),
		}

		childElement.SetBounds(ui.Bounds{
			Position: ui.NewPosition(-int(e.offsetX), -int(e.offsetY)),
			Size:     childSize,
		})
	}

	e.maxOffsetX = float64(maxInt(0, maxChildSize.Width-contentBounds.Width))
	e.maxOffsetY = float64(maxInt(0, maxChildSize.Height-contentBounds.Height))
	e.offsetX = dprec.Clamp(e.offsetX, 0.0, e.maxOffsetX)
	e.offsetY = dprec.Clamp(e.offsetY, 0.0, e.maxOffsetY)

	element.SetIdealSize(ui.Size{
		Width:  maxChildSize.Width + element.Padding().Horizontal(),
		Height: maxChildSize.Height + element.Padding().Vertical(),
	})
}

func (e *scrollPaneEssence) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Code {
	case ui.KeyCodePageDown:
		if event.Type == ui.KeyboardEventTypeKeyDown || event.Type == ui.KeyboardEventTypeRepeat {
			e.scroll(element, 0.0, -100.0)
			return true
		}
	case ui.KeyCodePageUp:
		if event.Type == ui.KeyboardEventTypeKeyDown || event.Type == ui.KeyboardEventTypeRepeat {
			e.scroll(element, 0.0, 100.0)
			return true
		}
	}
	return false
}

func (e *scrollPaneEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	// TODO: Support mouse dragging as a means to scroll
	e.scroll(element, event.ScrollX*10.0, event.ScrollY*10.0)
	return true
}

func (e *scrollPaneEssence) scroll(element *ui.Element, deltaX, deltaY float64) {
	e.offsetX -= deltaX
	e.offsetY -= deltaY
	if e.scrollHorizontally && !e.scrollVertically {
		e.offsetX -= deltaY * 10
	}
	e.offsetX = dprec.Clamp(e.offsetX, 0.0, e.maxOffsetX)
	e.offsetY = dprec.Clamp(e.offsetY, 0.0, e.maxOffsetY)

	e.Apply(element)
	element.Invalidate()
}
