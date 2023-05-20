package mat

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

type LabelData struct {
	Font                *ui.Font
	FontSize            opt.T[float32]
	FontColor           opt.T[ui.Color]
	HorizontalAlignment layout.HorizontalAlignment
	VerticalAlignment   layout.VerticalAlignment
	Text                string
}

var Label = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data LabelData
	)
	props.InjectOptionalData(&data, LabelData{})

	essence := co.UseState(func() *labelEssence {
		return &labelEssence{}
	}).Get()

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

	txtSize := essence.font.TextSize(essence.text, essence.fontSize)

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence:   essence,
			IdealSize: opt.V(ui.NewSize(int(txtSize.X), int(txtSize.Y))),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementRenderHandler = (*labelEssence)(nil)

type labelEssence struct {
	font      *ui.Font
	fontSize  float32
	fontColor ui.Color
	text      string
}

func (b *labelEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	if b.font != nil && b.text != "" {
		contentArea := element.ContentBounds()
		textDrawSize := b.font.TextSize(b.text, b.fontSize)

		canvas.Reset()
		// TODO: Take horizontal and vertical alignment into account.
		canvas.FillText(b.text, sprec.NewVec2(
			float32(contentArea.X)+(float32(contentArea.Width)-textDrawSize.X)/2,
			float32(contentArea.Y)+(float32(contentArea.Height)-textDrawSize.Y)/2,
		), ui.Typography{
			Font:  b.font,
			Size:  b.fontSize,
			Color: b.fontColor,
		})
	}
}
