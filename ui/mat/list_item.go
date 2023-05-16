package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mat/layout"
)

var (
	ListItemPadding = 2
)

// ListItemData holds the data for a ListItem component.
type ListItemData struct {
	Selected bool
}

var defaultListItemData = ListItemData{}

// ListItemCallbackData holds the callback data for a ListItem component.
type ListItemCallbackData struct {
	OnSelected func()
}

var defaultListItemCallbackData = ListItemCallbackData{
	OnSelected: func() {},
}

// ListItem represents a component to be displayed in a List.
var ListItem = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetOptionalData(props, defaultListItemData)
		callbackData = co.GetOptionalCallbackData(props, defaultListItemCallbackData)
	)

	essence := co.UseState(func() *listItemEssence {
		return &listItemEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.OnSelected),
		}
	}).Get()
	essence.selected = data.Selected

	return co.New(Element, func() {
		co.WithData(ElementData{
			Padding: ui.Spacing{
				Left:   ListItemPadding,
				Right:  ListItemPadding,
				Top:    ListItemPadding,
				Bottom: ListItemPadding,
			},
			Essence: essence,
			Layout:  layout.Fill(),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementRenderHandler = (*listItemEssence)(nil)

type listItemEssence struct {
	*ButtonBaseEssence

	selected bool
}

func (e *listItemEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch e.State() {
	case ButtonStateOver:
		backgroundColor = HoverOverlayColor
	case ButtonStateDown:
		backgroundColor = PressOverlayColor
	default:
		backgroundColor = ui.Transparent()
	}
	if e.selected {
		backgroundColor = SecondaryColor
	}

	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	canvas.Reset()
	canvas.SetStrokeSize(1.0)
	canvas.SetStrokeColor(OutlineColor)
	canvas.Rectangle(
		sprec.ZeroVec2(),
		sprec.NewVec2(width, height),
	)
	if !backgroundColor.Transparent() {
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	canvas.Stroke()
}
