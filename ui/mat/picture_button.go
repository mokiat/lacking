package mat

import (
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/optional"
	t "github.com/mokiat/lacking/ui/template"
)

type PictureButtonData struct {
	Font      ui.Font
	FontSize  optional.Int
	FontColor optional.Color
	UpImage   ui.Image
	OverImage ui.Image
	DownImage ui.Image
	Text      string
}

type PictureButtonCallbackData struct {
	ClickListener ClickListener
}

var PictureButton = t.ShallowCached(t.Plain(func(props t.Properties) t.Instance {
	var (
		data         PictureButtonData
		callbackData PictureButtonCallbackData
		essence      *pictureButtonEssence
	)
	props.InjectData(&data)
	props.InjectCallbackData(&callbackData)

	t.UseState(func() interface{} {
		return &pictureButtonEssence{
			state:     buttonStateUp,
			fontSize:  24,
			fontColor: ui.Black(),
		}
	}).Inject(&essence)

	essence.font = data.Font
	if data.FontSize.Specified {
		essence.fontSize = data.FontSize.Value
	}
	if data.FontColor.Specified {
		essence.fontColor = data.FontColor.Value
	}
	essence.upImage = data.UpImage
	essence.overImage = data.OverImage
	essence.downImage = data.DownImage
	essence.text = data.Text
	essence.clickListener = callbackData.ClickListener

	return t.New(Element, func() {
		t.WithData(ElementData{
			Essence: essence,
		})
		t.WithLayoutData(props.LayoutData())
		t.WithChildren(props.Children())
	})
}))

var _ ui.ElementMouseHandler = (*pictureButtonEssence)(nil)
var _ ui.ElementRenderHandler = (*pictureButtonEssence)(nil)

type pictureButtonEssence struct {
	font      ui.Font
	fontSize  int
	fontColor ui.Color
	text      string
	upImage   ui.Image
	overImage ui.Image
	downImage ui.Image

	clickListener ClickListener

	state buttonState
}

func (e *pictureButtonEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	context := element.Context()
	switch event.Type {
	case ui.MouseEventTypeEnter:
		e.state = buttonStateOver
		context.Window().Invalidate()
	case ui.MouseEventTypeLeave:
		e.state = buttonStateUp
		context.Window().Invalidate()
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if e.state == buttonStateDown {
				e.onClick()
			}
			e.state = buttonStateOver
			context.Window().Invalidate()
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			e.state = buttonStateDown
			context.Window().Invalidate()
		}
	}
	return true
}

func (e *pictureButtonEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	var visibleImage ui.Image
	switch e.state {
	case buttonStateUp:
		visibleImage = e.upImage
	case buttonStateOver:
		visibleImage = e.overImage
	case buttonStateDown:
		visibleImage = e.downImage
	}
	if visibleImage != nil {
		canvas.DrawImage(visibleImage,
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	} else {
		canvas.SetSolidColor(ui.Black())
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	}
	if e.font != nil && e.text != "" {
		canvas.SetFont(e.font)
		canvas.SetFontSize(e.fontSize)
		canvas.SetSolidColor(e.fontColor)

		contentArea := element.ContentBounds()
		textDrawSize := canvas.TextSize(e.text)
		canvas.DrawText(e.text, ui.NewPosition(
			contentArea.X+(contentArea.Width-textDrawSize.Width)/2,
			contentArea.Y+(contentArea.Height-textDrawSize.Height)/2,
		))
	}
}

func (e *pictureButtonEssence) onClick() {
	if e.clickListener != nil {
		e.clickListener()
	}
}
