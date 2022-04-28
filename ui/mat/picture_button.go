package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/optional"
)

type PictureButtonData struct {
	Font      *ui.Font
	FontSize  optional.Float32
	FontColor optional.Color
	UpImage   *ui.Image
	OverImage *ui.Image
	DownImage *ui.Image
	Padding   ui.Spacing
	Text      string
}

type PictureButtonCallbackData struct {
	ClickListener ClickListener
}

var PictureButton = co.ShallowCached(co.Define(func(props co.Properties) co.Instance {
	var (
		data         PictureButtonData
		callbackData PictureButtonCallbackData
		essence      *pictureButtonEssence
	)
	props.InjectOptionalData(&data, PictureButtonData{})
	props.InjectOptionalCallbackData(&callbackData, PictureButtonCallbackData{})

	co.UseState(func() interface{} {
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

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Padding: data.Padding,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
}))

var _ ui.ElementMouseHandler = (*pictureButtonEssence)(nil)
var _ ui.ElementRenderHandler = (*pictureButtonEssence)(nil)

type pictureButtonEssence struct {
	font      *ui.Font
	fontSize  float32
	fontColor ui.Color
	text      string
	upImage   *ui.Image
	overImage *ui.Image
	downImage *ui.Image

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

func (e *pictureButtonEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var visibleImage *ui.Image
	switch e.state {
	case buttonStateUp:
		visibleImage = e.upImage
	case buttonStateOver:
		visibleImage = e.overImage
	case buttonStateDown:
		visibleImage = e.downImage
	}
	if visibleImage != nil {
		canvas.Shape().Begin(ui.Fill{
			Color:       ui.White(),
			Image:       visibleImage,
			ImageOffset: sprec.ZeroVec2(),
			ImageSize: sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		})
		canvas.Shape().Rectangle(
			sprec.ZeroVec2(),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Shape().End()
	} else {
		canvas.Shape().Begin(ui.Fill{
			Color: ui.Black(),
		})
		canvas.Shape().Rectangle(
			sprec.ZeroVec2(),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Shape().End()
	}
	if e.font != nil && e.text != "" {
		canvas.Text().Begin(ui.Typography{
			Font:  e.font,
			Size:  e.fontSize,
			Color: e.fontColor,
		})
		contentArea := element.ContentBounds()
		textDrawSize := e.font.TextSize(e.text, e.fontSize)
		canvas.Text().Line(e.text, sprec.NewVec2(
			float32(contentArea.X)+(float32(contentArea.Width)-textDrawSize.X)/2,
			float32(contentArea.Y)+(float32(contentArea.Height)-textDrawSize.Y)/2,
		))
		canvas.Text().End()
	}
}

func (e *pictureButtonEssence) onClick() {
	if e.clickListener != nil {
		e.clickListener()
	}
}
