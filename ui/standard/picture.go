package standard

import (
	"fmt"

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
	result := &picture{}

	element := ctx.CreateElement()
	element.SetLayoutConfig(layoutConfig)
	element.SetHandler(result)

	result.Control = ctx.CreateControl(element)
	element.SetControl(result)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}

	return result, nil
}

var _ ui.ElementRenderHandler = (*picture)(nil)

type picture struct {
	ui.Control

	image ui.Image
}

func (p *picture) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := p.Element().ApplyAttributes(attributes); err != nil {
		return err
	}
	if src, ok := attributes.StringAttribute("src"); ok {
		context := p.Element().Context()
		img, err := context.OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open image: %w", err)
		}
		p.image = img
	}
	return nil
}

func (p *picture) OnRender(element *ui.Element, canvas ui.Canvas) {
	if p.image != nil {
		canvas.DrawImage(
			p.image,
			ui.NewPosition(0, 0),
			element.Bounds().Size,
		)
	}
}
