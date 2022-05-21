package mat

import (
	"github.com/mokiat/gomath/sprec"
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

var ScrollPane = co.Define(func(props co.Properties) co.Instance {
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
			Essence: essence,
			Layout:  essence,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.Layout = (*scrollPaneEssence)(nil)
var _ ui.ElementMouseHandler = (*scrollPaneEssence)(nil)
var _ ui.ElementRenderHandler = (*scrollPaneEssence)(nil)

type scrollPaneEssence struct {
	scrollHorizontally bool
	scrollVertically   bool

	offsetX float64
	offsetY float64
}

func (l *scrollPaneEssence) Apply(element *ui.Element) {
	contentBounds := element.ContentBounds()
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if !l.scrollHorizontally && layoutConfig.GrowHorizontally {
			childSize.Width = maxInt(childSize.Width, contentBounds.Width)
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}
		if !l.scrollVertically && layoutConfig.GrowVertically {
			childSize.Height = maxInt(childSize.Height, contentBounds.Height)
		}

		childElement.SetBounds(ui.Bounds{
			Position: ui.NewPosition(int(l.offsetX), int(l.offsetY)),
			Size:     childSize,
		})
	}

	element.SetIdealSize(l.calculateIdealSize(element))
}

func (e *scrollPaneEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Type {
	case ui.MouseEventTypeScroll:
		e.offsetX += event.ScrollX * 10
		e.offsetY += event.ScrollY * 10
		e.Apply(element)
		element.Context().Window().Invalidate()
		return true
	default:
		return false
	}
}

func (e *scrollPaneEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	canvas.Reset()
	canvas.Rectangle(
		sprec.NewVec2(0, 0),
		sprec.NewVec2(
			float32(element.Bounds().Width),
			float32(element.Bounds().Height),
		),
	)
	canvas.Fill(ui.Fill{
		Color: ui.RGB(30, 255, 128),
	})
}

func (l *scrollPaneEssence) calculateIdealSize(element *ui.Element) ui.Size {
	result := ui.NewSize(0, 0)
	for childElement := element.FirstChild(); childElement != nil; childElement = childElement.RightSibling() {
		layoutConfig := ElementLayoutData(childElement)

		childSize := childElement.IdealSize()
		if layoutConfig.Width.Specified {
			childSize.Width = layoutConfig.Width.Value
		}
		if layoutConfig.Height.Specified {
			childSize.Height = layoutConfig.Height.Value
		}

		result.Width = maxInt(result.Width, childSize.Width)
		result.Height = maxInt(result.Height, childSize.Height)
	}
	result.Width += element.Padding().Horizontal()
	result.Height += element.Padding().Vertical()
	return result
}
