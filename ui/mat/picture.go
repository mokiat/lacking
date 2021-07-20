package mat

import (
	"github.com/mokiat/lacking/ui"
	t "github.com/mokiat/lacking/ui/template"
)

// PictureData represents the available data properties for the
// Picture component.
type PictureData struct {

	// BackgroundColor specifies the color to be rendered behind the image.
	BackgroundColor ui.Color

	// Image specifies the Image to be displayed.
	Image ui.Image

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

var Picture = t.ShallowCached(t.Define(func(props t.Properties) t.Instance {
	var data PictureData
	props.InjectData(&data)

	return t.New(Element, func() {
		t.WithData(t.ElementData{
			Essence: &pictureEssence{
				PictureData: data,
			},
		})
		t.WithLayoutData(props.LayoutData())
		t.WithChildren(props.Children())
	})
}))

var _ ui.ElementRenderHandler = (*pictureEssence)(nil)

type pictureEssence struct {
	PictureData
}

func (p *pictureEssence) OnRender(element *ui.Element, canvas ui.Canvas) {
	if !p.BackgroundColor.Transparent() {
		canvas.SetSolidColor(p.BackgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	}

	if p.Image != nil {
		drawPosition, drawSize := p.evaluateImageDrawLocation(element)
		canvas.DrawImage(
			p.Image,
			drawPosition,
			drawSize,
		)
	}
}

func (p *pictureEssence) evaluateImageDrawLocation(element *ui.Element) (ui.Position, ui.Size) {
	elementSize := element.Bounds().Size
	imageSize := p.Image.Size()
	determinant := imageSize.Width*elementSize.Height - imageSize.Height*elementSize.Width
	imageHasHigherAspectRatio := determinant >= 0

	switch p.Mode {
	case ImageModeCover:
		if imageHasHigherAspectRatio {
			return ui.NewPosition(
					-determinant/(2*imageSize.Height),
					0,
				),
				ui.NewSize(
					(elementSize.Height*imageSize.Width)/imageSize.Height,
					elementSize.Height,
				)
		} else {
			return ui.NewPosition(
					0,
					determinant/(2*imageSize.Width),
				),
				ui.NewSize(
					elementSize.Width,
					(elementSize.Width*imageSize.Height)/imageSize.Width,
				)
		}

	case ImageModeFit:
		if imageHasHigherAspectRatio {
			return ui.NewPosition(
					0,
					determinant/(2*imageSize.Width),
				),
				ui.NewSize(
					elementSize.Width,
					(elementSize.Width*imageSize.Height)/imageSize.Width,
				)
		} else {
			return ui.NewPosition(
					-determinant/(2*imageSize.Height),
					0,
				),
				ui.NewSize(
					(elementSize.Height*imageSize.Width)/imageSize.Height,
					elementSize.Height,
				)
		}

	default:
		return ui.NewPosition(0, 0), elementSize
	}
}
