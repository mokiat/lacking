package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

var (
	DropdownFontSize = float32(18)
	DropdownFontFile = "mat:///roboto-regular.ttf"
	DropdownIconSize = 24
	DropdownIconFile = "mat:///expanded.png"
)

// DropdownData holds the data for a Dropdown container.
type DropdownData struct {
	Items       []DropdownItem
	SelectedKey interface{}
}

var defaultDropdownData = DropdownData{}

// DropdownItem represents a single dropdown entry.
type DropdownItem struct {
	Key   interface{}
	Label string
}

// DropdownCallbackData holds the callback data for a Dropdown container.
type DropdownCallbackData struct {
	OnItemSelected func(key interface{})
}

var defaultDropdownCallbackData = DropdownCallbackData{
	OnItemSelected: func(key interface{}) {},
}

// Dropdown is a container that hides a number of UI options in a compact way.
var Dropdown = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetOptionalData(props, defaultDropdownData)
		callbackData = co.GetOptionalCallbackData(props, defaultDropdownCallbackData)
	)

	listOverlay := co.UseState(func() co.Overlay {
		return nil
	})

	onClose := func() bool {
		overlay := listOverlay.Get()
		overlay.Close()
		listOverlay.Set(nil)
		return true
	}

	onItemSelected := func(key interface{}) {
		onClose()
		callbackData.OnItemSelected(key)
	}

	onOpen := func() {
		overlay := co.OpenOverlay(co.New(dropdownList, func() {
			co.WithData(data)
			co.WithCallbackData(dropdownListCallbackData{
				OnSelected: onItemSelected,
				OnClose:    onClose,
			})
		}))
		listOverlay.Set(overlay)
	}

	essence := co.UseState(func() *dropdownEssence {
		return &dropdownEssence{
			ButtonBaseEssence: NewButtonBaseEssence(onOpen),
		}
	}).Get()
	essence.SetOnClick(onOpen) // override onOpen as old one refs old data

	label := ""
	for _, item := range data.Items {
		if item.Key == data.SelectedKey {
			label = item.Label
		}
	}

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Layout:  NewAnchorLayout(AnchorLayoutSettings{}),
		})
		co.WithLayoutData(props.LayoutData())

		co.WithChild("label", co.New(Label, func() {
			co.WithData(LabelData{
				Font:      co.OpenFont(scope, DropdownFontFile),
				FontSize:  optional.Value(DropdownFontSize),
				FontColor: optional.Value(OnSurfaceColor),
				Text:      label,
			})
			co.WithLayoutData(LayoutData{
				Left:           optional.Value(0),
				Right:          optional.Value(DropdownIconSize),
				VerticalCenter: optional.Value(0),
			})
		}))

		co.WithChild("button", co.New(Picture, func() {
			co.WithData(PictureData{
				Image:      co.OpenImage(scope, DropdownIconFile),
				ImageColor: optional.Value(OnSurfaceColor),
				Mode:       ImageModeFit,
			})
			co.WithLayoutData(LayoutData{
				Width:          optional.Value(DropdownIconSize),
				Height:         optional.Value(DropdownIconSize),
				Right:          optional.Value(0),
				VerticalCenter: optional.Value(0),
			})
		}))
	})
})

var _ ui.ElementRenderHandler = (*dropdownEssence)(nil)

type dropdownEssence struct {
	*ButtonBaseEssence
}

func (e *dropdownEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch e.State() {
	case ButtonStateOver:
		backgroundColor = HoverOverlayColor
	case ButtonStateDown:
		backgroundColor = PressOverlayColor
	default:
		backgroundColor = ui.Transparent()
	}

	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	canvas.Reset()
	canvas.SetStrokeSize(2.0)
	canvas.SetStrokeColor(PrimaryLightColor)
	canvas.RoundRectangle(
		sprec.ZeroVec2(),
		sprec.NewVec2(width, height),
		sprec.NewVec4(5.0, 5.0, 5.0, 5.0),
	)
	if !backgroundColor.Transparent() {
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	canvas.Stroke()
}
