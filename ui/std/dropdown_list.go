package std

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

var (
	DropdownListFontSize = float32(18)
	DropdownListFontFile = "ui:///roboto-regular.ttf"
)

type dropdownListCallbackData struct {
	OnSelected func(key any)
	OnClose    OnActionFunc
}

var dropdownList = co.DefineType(&dropdownListComponent{})

type dropdownListComponent struct {
	Scope      co.Scope      `co:"scope"`
	Properties co.Properties `co:"properties"`

	items           []DropdownItem
	selectedItemKey any

	onSelected func(key any)
	onClose    OnActionFunc
}

func (c *dropdownListComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, dropdownDefaultData)
	c.items = data.Items
	c.selectedItemKey = data.SelectedKey

	callbackData := co.GetCallbackData[dropdownListCallbackData](c.Properties)
	c.onSelected = callbackData.OnSelected
	c.onClose = callbackData.OnClose
}

func (c *dropdownListComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithData(co.ElementData{
			Essence: c,
			Layout:  layout.Anchor(),
		})
		co.WithLayoutData(c.Properties.LayoutData())

		co.WithChild("content", co.New(Paper, func() {
			co.WithLayoutData(layout.Data{
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
			co.WithData(PaperData{
				Layout: layout.Vertical(layout.VerticalSettings{
					ContentSpacing: 5,
				}),
			})

			for i, item := range c.items {
				func(i int, item DropdownItem) {
					co.WithChild(fmt.Sprintf("item-%d", i), co.New(dropdownItem, func() {
						co.WithLayoutData(layout.Data{
							GrowHorizontally: true,
						})
						co.WithData(dropdownItemData{
							Selected: item.Key == c.selectedItemKey,
						})
						co.WithCallbackData(dropdownItemCallbackData{
							OnSelected: func() {
								c.onSelected(item.Key)
							},
						})
						co.WithChild("label", co.New(Label, func() {
							co.WithData(LabelData{
								Font:      co.OpenFont(c.Scope, DropdownListFontFile),
								FontSize:  opt.V(DropdownListFontSize),
								FontColor: opt.V(OnSurfaceColor),
								Text:      item.Label,
							})
						}))
					}))
				}(i, item)
			}
		}))
	})
}

func (c *dropdownListComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	if event.Type == ui.MouseEventTypeUp && event.Button == ui.MouseButtonLeft {
		c.onClose()
		return true
	}
	return false
}

func (c *dropdownListComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
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
