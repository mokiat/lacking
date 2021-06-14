package standard

import (
	"fmt"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("PictureButton", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildPictureButton(ctx, template, layoutConfig)
	}))
}

// PictureButton represets a Button control that uses images
// for its various states.
type PictureButton interface {
	ui.Control

	// SetClickListener registers the following listener to
	// be called when this button is clicked.
	SetClickListener(clickListener PictureButtonClickListener)
}

type PictureButtonClickListener func(button PictureButton)

// BuildPictureButton constructs a new PictureButton control.
func BuildPictureButton(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (PictureButton, error) {
	result := &pictureButton{
		state: buttonStateUp,
	}

	element := ctx.CreateElement()
	element.SetLayoutConfig(layoutConfig)
	element.SetHandler(result)

	result.Control = ctx.CreateControl(element)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}

	return result, nil
}

var _ ui.ElementMouseHandler = (*pictureButton)(nil)
var _ ui.ElementRenderHandler = (*pictureButton)(nil)

type pictureButton struct {
	ui.Control

	state buttonState

	upImage   ui.Image
	overImage ui.Image
	downImage ui.Image

	clickListener PictureButtonClickListener
}

func (b *pictureButton) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := b.Control.ApplyAttributes(attributes); err != nil {
		return err
	}
	if src, ok := attributes.StringAttribute("src-up"); ok {
		img, err := b.Context().OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open 'up' image: %w", err)
		}
		b.upImage = img
	}
	if src, ok := attributes.StringAttribute("src-ovar"); ok {
		img, err := b.Context().OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open 'over' image: %w", err)
		}
		b.overImage = img
	}
	if src, ok := attributes.StringAttribute("src-down"); ok {
		img, err := b.Context().OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open 'down' image: %w", err)
		}
		b.downImage = img
	}
	return nil
}

func (b *pictureButton) SetClickListener(listener PictureButtonClickListener) {
	b.clickListener = listener
}

func (b *pictureButton) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Type {
	case ui.MouseEventTypeEnter:
		b.state = buttonStateOver
		b.Context().Window().Invalidate()
	case ui.MouseEventTypeLeave:
		b.state = buttonStateUp
		b.Context().Window().Invalidate()
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if b.state == buttonStateDown && b.clickListener != nil {
				b.clickListener(b)
			}
			b.state = buttonStateOver
			b.Context().Window().Invalidate()
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			b.state = buttonStateDown
			b.Context().Window().Invalidate()
		}
	}
	return true
}

func (b *pictureButton) OnRender(element *ui.Element, canvas ui.Canvas) {
	var visibleImage ui.Image
	switch b.state {
	case buttonStateUp:
		visibleImage = b.upImage
	case buttonStateOver:
		visibleImage = b.overImage
	case buttonStateDown:
		visibleImage = b.downImage
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
}