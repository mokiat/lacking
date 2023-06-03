package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

var (
	LabelFontFile = "ui:///roboto-regular.ttf"
	LabelFontSize = float32(24.0)
)

// LabelData holds the data for the Label component.
type LabelData struct {
	Font      *ui.Font
	FontSize  opt.T[float32]
	FontColor opt.T[ui.Color]
	Text      string
}

var labelDefaultData = LabelData{}

// Label represents a component that visualizes a text string.
var Label = co.Define(&labelComponent{})

type labelComponent struct {
	co.BaseComponent

	font      *ui.Font
	fontSize  float32
	fontColor ui.Color
	text      string
}

func (c *labelComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), labelDefaultData)
	if data.Font != nil {
		c.font = data.Font
	} else {
		c.font = co.OpenFont(c.Scope(), "ui:///roboto-regular.ttf")
	}
	if data.FontSize.Specified {
		c.fontSize = data.FontSize.Value
	} else {
		c.fontSize = LabelFontSize
	}
	if data.FontColor.Specified {
		c.fontColor = data.FontColor.Value
	} else {
		c.fontColor = OnSurfaceColor
	}
	c.text = data.Text
}

func (c *labelComponent) Render() co.Instance {
	textSize := c.font.TextSize(c.text, c.fontSize)
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence:   c,
			IdealSize: opt.V(ui.NewSize(int(textSize.X), int(textSize.Y))),
		})
		co.WithChildren(c.Properties().Children())
	})
}

func (c *labelComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	if c.text != "" {
		drawBounds := canvas.DrawBounds(element, false)
		textDrawSize := c.font.TextSize(c.text, c.fontSize)

		canvas.Reset()
		canvas.FillText(c.text, sprec.NewVec2(
			(drawBounds.Width()-textDrawSize.X)/2.0,
			(drawBounds.Height()-textDrawSize.Y)/2.0,
		), ui.Typography{
			Font:  c.font,
			Size:  c.fontSize,
			Color: c.fontColor,
		})
	}
}
