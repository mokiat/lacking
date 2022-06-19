package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

// PictureData represents the available data properties for the
// Picture component.
type PictureData struct {

	// BackgroundColor specifies the color to be rendered behind the image.
	BackgroundColor optional.V[ui.Color]

	// Image specifies the Image to be displayed.
	Image *ui.Image

	// ImageColor specifies an optional multiplier color.
	ImageColor optional.V[ui.Color]

	// Mode specifies how the image will be scaled and visualized within the
	// Picture component.
	Mode ImageMode
}

// ImageMode determines how an image is displayed within a
// Picture control.
type ImageMode int

const (
	// ImageModeStretch will stretch the image to cover the
	// available draw area in the control.
	ImageModeStretch ImageMode = 1 + iota

	// ImageModeFit will preserve the aspect ratio of the image
	// and will scale it up or down so that the image takes as
	// much space of the draw area as possible, without exiting
	// the bounds of the area.
	ImageModeFit

	// ImageModeCover will preserve the aspect ratio of the image
	// while also ensuring it covers the entire draw area. This
	// usually means that part of the image will be outside the
	// bounds of the control and will not be visible.
	ImageModeCover
)

var Picture = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data PictureData
	)
	props.InjectOptionalData(&data, PictureData{})

	essence := co.UseState(func() *pictureEssence {
		return &pictureEssence{}
	}).Get()

	if data.BackgroundColor.Specified {
		essence.backgroundColor = data.BackgroundColor.Value
	} else {
		essence.backgroundColor = ui.Transparent()
	}
	if data.ImageColor.Specified {
		essence.imageColor = data.ImageColor.Value
	} else {
		essence.imageColor = ui.White()
	}
	essence.image = data.Image
	essence.mode = data.Mode

	var idealSize optional.V[ui.Size]
	if data.Image != nil {
		idealSize = optional.Value(data.Image.Size())
	}

	return co.New(Element, func() {
		co.WithData(co.ElementData{
			Essence:   essence,
			IdealSize: idealSize,
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementRenderHandler = (*pictureEssence)(nil)

type pictureEssence struct {
	backgroundColor ui.Color
	image           *ui.Image
	imageColor      ui.Color
	mode            ImageMode
}

func (p *pictureEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	if !p.backgroundColor.Transparent() {
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Fill(ui.Fill{
			Color: p.backgroundColor,
		})
	}

	if p.image != nil {
		drawPosition, drawSize := p.evaluateImageDrawLocation(element)
		canvas.Reset()
		canvas.Rectangle(
			drawPosition,
			drawSize,
		)
		canvas.Fill(ui.Fill{
			Color:       p.imageColor,
			Image:       p.image,
			ImageOffset: drawPosition,
			ImageSize:   drawSize,
		})
	}
}

func (p *pictureEssence) evaluateImageDrawLocation(element *ui.Element) (sprec.Vec2, sprec.Vec2) {
	elementSize := element.Bounds().Size
	imageSize := p.image.Size()
	determinant := imageSize.Width*elementSize.Height - imageSize.Height*elementSize.Width
	imageHasHigherAspectRatio := determinant >= 0

	switch p.mode {
	case ImageModeCover:
		if imageHasHigherAspectRatio {
			return sprec.NewVec2(
					-float32(determinant/(2*imageSize.Height)),
					0,
				),
				sprec.NewVec2(
					float32((elementSize.Height*imageSize.Width)/imageSize.Height),
					float32(elementSize.Height),
				)
		} else {
			return sprec.NewVec2(
					0,
					float32(determinant/(2*imageSize.Width)),
				),
				sprec.NewVec2(
					float32(elementSize.Width),
					float32((elementSize.Width*imageSize.Height)/imageSize.Width),
				)
		}

	case ImageModeFit:
		if imageHasHigherAspectRatio {
			return sprec.NewVec2(
					0,
					float32(determinant/(2*imageSize.Width)),
				),
				sprec.NewVec2(
					float32(elementSize.Width),
					float32((elementSize.Width*imageSize.Height)/imageSize.Width),
				)
		} else {
			return sprec.NewVec2(
					-float32(determinant/(2*imageSize.Height)),
					0,
				),
				sprec.NewVec2(
					float32((elementSize.Height*imageSize.Width)/imageSize.Height),
					float32(elementSize.Height),
				)
		}

	default:
		return sprec.ZeroVec2(), sprec.NewVec2(float32(elementSize.Width), float32(elementSize.Height))
	}
}
