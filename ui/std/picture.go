package std

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

// ImageMode determines how an image is displayed within a
// Picture control.
type ImageMode int

const (
	// ImageModeStretch will stretch the image to cover the
	// available draw area in the control.
	ImageModeStretch ImageMode = iota

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

// PictureData represents the available data properties for the
// Picture component.
type PictureData struct {

	// BackgroundColor specifies the color to be rendered behind the image.
	BackgroundColor opt.T[ui.Color]

	// Image specifies the Image to be displayed.
	Image *ui.Image

	// ImageColor specifies an optional multiplier color.
	ImageColor opt.T[ui.Color]

	// Mode specifies how the image will be scaled and visualized within the
	// Picture component.
	Mode ImageMode
}

var pictureDefaultData = PictureData{}

var Picture = co.Define(&pictureComponent{})

type pictureComponent struct {
	Properties co.Properties `co:"properties"`

	backgroundColor ui.Color
	image           *ui.Image
	imageColor      ui.Color
	mode            ImageMode
}

func (c *pictureComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties, pictureDefaultData)
	if data.BackgroundColor.Specified {
		c.backgroundColor = data.BackgroundColor.Value
	} else {
		c.backgroundColor = ui.Transparent()
	}
	c.image = data.Image
	if data.ImageColor.Specified {
		c.imageColor = data.ImageColor.Value
	} else {
		c.imageColor = ui.White() // full mask
	}
	c.mode = data.Mode
}

func (c *pictureComponent) Render() co.Instance {
	var idealSize opt.T[ui.Size]
	if c.image != nil {
		idealSize = opt.V(c.image.Size())
	}
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties.LayoutData())
		co.WithData(co.ElementData{
			Essence:   c,
			IdealSize: idealSize,
		})
		co.WithChildren(c.Properties.Children())
	})
}

func (c *pictureComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	if !c.backgroundColor.Transparent() {
		canvas.Reset()
		canvas.Rectangle(
			sprec.NewVec2(0, 0),
			sprec.NewVec2(
				float32(element.Bounds().Size.Width),
				float32(element.Bounds().Size.Height),
			),
		)
		canvas.Fill(ui.Fill{
			Color: c.backgroundColor,
		})
	}

	if c.image != nil {
		drawPosition, drawSize := c.evaluateImageDrawLocation(element)
		canvas.Reset()
		canvas.Rectangle(
			drawPosition,
			drawSize,
		)
		canvas.Fill(ui.Fill{
			Color:       c.imageColor,
			Image:       c.image,
			ImageOffset: drawPosition,
			ImageSize:   drawSize,
		})
	}
}

func (c *pictureComponent) evaluateImageDrawLocation(element *ui.Element) (sprec.Vec2, sprec.Vec2) {
	elementSize := element.Bounds().Size
	imageSize := c.image.Size()
	determinant := imageSize.Width*elementSize.Height - imageSize.Height*elementSize.Width
	imageHasHigherAspectRatio := determinant >= 0

	switch c.mode {
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

	case ImageModeStretch:
		fallthrough

	default:
		return sprec.ZeroVec2(), sprec.NewVec2(float32(elementSize.Width), float32(elementSize.Height))
	}
}
