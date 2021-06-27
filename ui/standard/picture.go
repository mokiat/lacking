package standard

import (
	"fmt"
	"strings"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("Picture", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildPicture(ctx, template, layoutConfig)
	}))
}

// Picture represents a Control that displays an Image.
type Picture interface {
	ui.Control
}

// BuildPicture constructs a new Picture control.
func BuildPicture(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (Picture, error) {
	result := &picture{
		element: ctx.CreateElement(),
		mode:    ImageModeStretch,
	}
	result.element.SetLayoutConfig(layoutConfig)
	result.element.SetEssence(result)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}
	return result, nil
}

var _ ui.ElementRenderHandler = (*picture)(nil)

type picture struct {
	element *ui.Element

	backgroundColor *ui.Color
	image           ui.Image
	mode            ImageMode
}

func (p *picture) Element() *ui.Element {
	return p.element
}

func (p *picture) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := p.element.ApplyAttributes(attributes); err != nil {
		return err
	}
	if color, ok := attributes.ColorAttribute("background-color"); ok {
		p.backgroundColor = &color
	}
	if src, ok := attributes.StringAttribute("src"); ok {
		context := p.element.Context()
		img, err := context.OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open image: %w", err)
		}
		p.image = img
	}
	if mode, ok := ImageModeAttribute(attributes, "mode"); ok {
		p.mode = mode
	}
	return nil
}

func (p *picture) OnRender(element *ui.Element, canvas ui.Canvas) {
	if p.backgroundColor != nil {
		canvas.SetSolidColor(*p.backgroundColor)
		canvas.FillRectangle(
			ui.NewPosition(0, 0),
			p.element.Bounds().Size,
		)
	}

	if p.image != nil {
		drawPosition, drawSize := p.evaluateImageDrawLocation()
		canvas.DrawImage(
			p.image,
			drawPosition,
			drawSize,
		)
	}
}

func (p *picture) evaluateImageDrawLocation() (ui.Position, ui.Size) {
	elementSize := p.element.Bounds().Size
	imageSize := p.image.Size()
	determinant := imageSize.Width*elementSize.Height - imageSize.Height*elementSize.Width
	imageHasHigherAspectRatio := determinant >= 0

	switch p.mode {
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

// AlignmentAttribute attempts to parse an Alignment from
// the attribute with the specified name.
func ImageModeAttribute(set ui.AttributeSet, name string) (ImageMode, bool) {
	if stringValue, ok := set.StringAttribute(name); ok {
		switch strings.ToLower(stringValue) {
		case "stretch":
			return ImageModeStretch, true
		case "fit":
			return ImageModeFit, true
		case "cover":
			return ImageModeCover, true
		}
	}
	return 0, false
}
