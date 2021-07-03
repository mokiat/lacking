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

// PictureButtonClickListener can be used to get notifications about
// picture button click events.
type PictureButtonClickListener func(button PictureButton)

// BuildPictureButton constructs a new PictureButton control.
func BuildPictureButton(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (PictureButton, error) {
	result := &pictureButton{
		element:   ctx.CreateElement(),
		state:     buttonStateUp,
		textSize:  24,
		textColor: ui.White(),
	}
	result.element.SetLayoutConfig(layoutConfig)
	result.element.SetEssence(result)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}
	return result, nil
}

var _ ui.ElementMouseHandler = (*pictureButton)(nil)
var _ ui.ElementRenderHandler = (*pictureButton)(nil)

type pictureButton struct {
	element *ui.Element

	state buttonState

	upImage   ui.Image
	overImage ui.Image
	downImage ui.Image
	font      ui.Font
	text      string
	textSize  int
	textColor ui.Color

	clickListener PictureButtonClickListener
}

func (b *pictureButton) Element() *ui.Element {
	return b.element
}

func (b *pictureButton) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := b.element.ApplyAttributes(attributes); err != nil {
		return err
	}
	context := b.element.Context()
	if src, ok := attributes.StringAttribute("src-up"); ok {
		img, err := context.OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open 'up' image: %w", err)
		}
		b.upImage = img
	}
	if src, ok := attributes.StringAttribute("src-over"); ok {
		img, err := context.OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open 'over' image: %w", err)
		}
		b.overImage = img
	}
	if src, ok := attributes.StringAttribute("src-down"); ok {
		img, err := context.OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open 'down' image: %w", err)
		}
		b.downImage = img
	}
	if familyStringValue, ok := attributes.StringAttribute("font-family"); ok {
		if subFamilyStringValue, ok := attributes.StringAttribute("font-style"); ok {
			font, found := context.GetFont(familyStringValue, subFamilyStringValue)
			if !found {
				return fmt.Errorf("could not find font %q / %q", familyStringValue, subFamilyStringValue)
			}
			b.font = font
		}
	}
	if stringValue, ok := attributes.StringAttribute("text"); ok {
		b.text = stringValue
	}
	if intValue, ok := attributes.IntAttribute("text-size"); ok {
		b.textSize = intValue
	}
	if colorValue, ok := attributes.ColorAttribute("text-color"); ok {
		b.textColor = colorValue
	}
	return nil
}

func (b *pictureButton) SetClickListener(listener PictureButtonClickListener) {
	b.clickListener = listener
}

func (b *pictureButton) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	context := b.element.Context()
	switch event.Type {
	case ui.MouseEventTypeEnter:
		b.state = buttonStateOver
		context.Window().Invalidate()
	case ui.MouseEventTypeLeave:
		b.state = buttonStateUp
		context.Window().Invalidate()
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if b.state == buttonStateDown && b.clickListener != nil {
				b.clickListener(b)
			}
			b.state = buttonStateOver
			context.Window().Invalidate()
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			b.state = buttonStateDown
			context.Window().Invalidate()
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
	if b.font != nil && b.text != "" {
		canvas.SetFont(b.font)
		canvas.SetFontSize(b.textSize)
		canvas.SetSolidColor(b.textColor)

		contentArea := element.ContentBounds()
		textDrawSize := canvas.TextSize(b.text)
		canvas.DrawText(b.text, ui.NewPosition(
			contentArea.X+(contentArea.Width-textDrawSize.Width)/2,
			contentArea.Y+(contentArea.Height-textDrawSize.Height)/2,
		))
	}
}
