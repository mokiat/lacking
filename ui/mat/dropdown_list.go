package mat

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

var (
	DropdownListFontSize = float32(18)
	DropdownListFontFile = "mat:///roboto-regular.ttf"
)

type dropdownListCallbackData struct {
	OnSelected func(key interface{})
	OnClose    func() bool
}

var dropdownList = co.Define(func(props co.Properties) co.Instance {
	var (
		data         = co.GetData[DropdownData](props)
		callbackData = co.GetCallbackData[dropdownListCallbackData](props)
	)

	essence := co.UseState(func() *dropdownListEssence {
		return &dropdownListEssence{
			onClose: callbackData.OnClose,
		}
	}).Get()

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Layout:  NewAnchorLayout(AnchorLayoutSettings{}),
		})

		co.WithChild("content", co.New(Paper, func() {
			co.WithData(PaperData{
				Layout: NewVerticalLayout(VerticalLayoutSettings{
					ContentSpacing: 5,
				}),
			})
			co.WithLayoutData(LayoutData{
				HorizontalCenter: optional.Value(0),
				VerticalCenter:   optional.Value(0),
			})

			for i, item := range data.Items {
				func(i int, item DropdownItem) {
					co.WithChild(fmt.Sprintf("item-%d", i), co.New(dropdownItem, func() {
						co.WithData(dropdownItemData{
							Selected: item.Key == data.SelectedKey,
						})
						co.WithLayoutData(LayoutData{
							GrowHorizontally: true,
						})
						co.WithCallbackData(dropdownItemCallbackData{
							OnSelected: func() {
								callbackData.OnSelected(item.Key)
							},
						})
						co.WithChild("label", co.New(Label, func() {
							co.WithData(LabelData{
								Font:      co.OpenFont(DropdownListFontFile),
								FontSize:  optional.Value(DropdownListFontSize),
								FontColor: optional.Value(OnSurfaceColor),
								Text:      item.Label,
							})
						}))
					}))
				}(i, item)
			}
		}))
	})
})

var _ ui.ElementMouseHandler = (*dropdownListEssence)(nil)
var _ ui.ElementRenderHandler = (*dropdownListEssence)(nil)

type dropdownListEssence struct {
	onClose func() bool
}

func (e *dropdownListEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	if event.Type == ui.MouseEventTypeUp && event.Button == ui.MouseButtonLeft {
		return e.onClose()
	}
	return false
}

func (e *dropdownListEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	canvas.Reset()
	canvas.Rectangle(
		sprec.ZeroVec2(),
		sprec.NewVec2(width, height),
	)
	canvas.Fill(ui.Fill{
		Color: ModalOverlayColor,
	})
}
