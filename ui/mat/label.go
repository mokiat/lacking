package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
	t "github.com/mokiat/lacking/ui/template"
)

type LabelData struct {
	Font      ui.Font
	FontSize  optional.Int
	FontColor optional.Color
	Text      string
}

var Label = t.ShallowCached(t.Plain(func(props t.Properties) t.Instance {
	var (
		data    LabelData
		essence *labelEssence
	)
	props.InjectData(&data)

	t.UseState(func() interface{} {
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

	return t.New(Element, func() {
		t.WithData(ElementData{
			Essence: essence,
		})
		t.WithLayoutData(props.LayoutData())
		t.WithChildren(props.Children())
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
		canvas.SetFont(b.font)
		canvas.SetFontSize(b.fontSize)
		canvas.SetSolidColor(b.fontColor)

		contentArea := element.ContentBounds()
		textDrawSize := canvas.TextSize(b.text)
		canvas.DrawText(b.text, ui.NewPosition(
			contentArea.X+(contentArea.Width-textDrawSize.Width)/2,
			contentArea.Y+(contentArea.Height-textDrawSize.Height)/2,
		))
	}
}
