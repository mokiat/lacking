package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/optional"
)

type LabelData struct {
	Font      ui.Font
	FontSize  optional.Int
	FontColor optional.Color
	Text      string
}

var Label = co.ShallowCached(co.Define(func(props co.Properties) co.Instance {
	var (
		data    LabelData
		essence *labelEssence
	)
	props.InjectOptionalData(&data, LabelData{})

	co.UseState(func() interface{} {
		return &labelEssence{}
	}).Inject(&essence)

	essence.font = data.Font
	if data.FontSize.Specified {
		essence.fontSize = data.FontSize.Value
	} else {
		essence.fontSize = 24
	}
	if data.FontColor.Specified {
		essence.fontColor = data.FontColor.Value
	} else {
		essence.fontColor = ui.Black()
	}
	essence.text = data.Text

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
}))

var _ ui.ElementRenderHandler = (*labelEssence)(nil)

type labelEssence struct {
	font      ui.Font
	fontSize  int
	fontColor ui.Color
	text      string
}

func (b *labelEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	if b.font != nil && b.text != "" {
		canvas.Text().Begin(ui.Typography{
			Font:  b.font,
			Size:  b.fontSize,
			Color: b.fontColor,
		})
		contentArea := element.ContentBounds()
		textDrawSize := b.font.TextSize(b.text, b.fontSize)
		canvas.Text().Line(b.text, ui.NewPosition(
			contentArea.X+(contentArea.Width-textDrawSize.Width)/2,
			contentArea.Y+(contentArea.Height-textDrawSize.Height)/2,
		))
		canvas.Text().End()
	}
}
