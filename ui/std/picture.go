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
	co.BaseComponent

	backgroundColor ui.Color
	image           *ui.Image
	imageColor      ui.Color
	mode            ImageMode
}

func (c *pictureComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), pictureDefaultData)
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
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(co.ElementData{
			Essence:   c,
			IdealSize: idealSize,
		})
		co.WithChildren(c.Properties().Children())
	})
}

func (c *pictureComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	drawBounds := canvas.DrawBounds(element, false)

	if !c.backgroundColor.Transparent() {
		canvas.Reset()
		canvas.Rectangle(
			drawBounds.Position,
			drawBounds.Size,
		)
		canvas.Fill(ui.Fill{
			Color: c.backgroundColor,
		})
	}

	if c.image != nil {
		pictureBounds := c.evaluateImageDrawLocation(drawBounds)
		canvas.Reset()
		canvas.Rectangle(
			pictureBounds.Position,
			pictureBounds.Size,
		)
		canvas.Fill(ui.Fill{
			Color:       c.imageColor,
			Image:       c.image,
			ImageOffset: pictureBounds.Position,
			ImageSize:   pictureBounds.Size,
		})
	}
}

func (c *pictureComponent) evaluateImageDrawLocation(drawBounds ui.DrawBounds) ui.DrawBounds {
	imageSize := sprec.NewVec2(
		float32(c.image.Size().Width),
		float32(c.image.Size().Height),
	)
	determinant := float32(imageSize.X)*drawBounds.Height() - float32(imageSize.Y)*drawBounds.Width()
	imageHasHigherAspectRatio := determinant >= 0

	switch c.mode {
	case ImageModeCover:
		if imageHasHigherAspectRatio {
			return ui.DrawBounds{
				Position: sprec.NewVec2(
					-determinant/(2.0*imageSize.Y),
					0.0,
				),
				Size: sprec.NewVec2(
					(drawBounds.Height()*imageSize.X)/imageSize.Y,
					drawBounds.Height(),
				),
			}
		} else {
			return ui.DrawBounds{
				Position: sprec.NewVec2(
					0.0,
					determinant/(2.0*imageSize.X),
				),
				Size: sprec.NewVec2(
					drawBounds.Width(),
					(drawBounds.Width()*imageSize.Y)/imageSize.X,
				),
			}
		}

	case ImageModeFit:
		if imageHasHigherAspectRatio {
			return ui.DrawBounds{
				Position: sprec.NewVec2(
					0.0,
					determinant/(2.0*imageSize.X),
				),
				Size: sprec.NewVec2(
					drawBounds.Width(),
					(drawBounds.Width()*imageSize.Y)/imageSize.X,
				),
			}
		} else {
			return ui.DrawBounds{
				Position: sprec.NewVec2(
					-determinant/(2.0*imageSize.Y),
					0,
				),
				Size: sprec.NewVec2(
					(drawBounds.Height()*imageSize.X)/imageSize.Y,
					drawBounds.Height(),
				),
			}
		}

	case ImageModeStretch:
		fallthrough

	default:
		return drawBounds
	}
}
