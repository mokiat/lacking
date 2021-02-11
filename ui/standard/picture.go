package standard

import (
	"fmt"
	"math/rand"

	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.Register("Picture", ui.BuilderFunc(func(ctx ui.BuildContext) (ui.Control, error) {
		return BuildPicture(ctx)
	}))
}

type Picture interface {
	ui.Control
}

func BuildPicture(ctx ui.BuildContext) (Picture, error) {
	element := ui.CreateElement()
	pic := &picture{
		Control: ui.CreateControl(
			ctx.Template.ID,
			element,
			ctx.Template.LayoutAttributes,
		),
		color: ui.RGBA(
			uint8(rand.Int()),
			uint8(rand.Int()),
			uint8(rand.Int()),
			255,
		),
	}
	element.SetControl(pic)
	element.SetHandler(pic)

	attributes := ctx.Template.Attributes
	if src, ok := attributes.StringAttribute("src"); ok {
		img, err := ctx.Window.OpenImage(src)
		if err != nil {
			return nil, fmt.Errorf("failed to open image: %w", err)
		}
		pic.image = img
	}

	return pic, nil
}

type picture struct {
	ui.Control
	color ui.Color
	image ui.Image
}

var _ ui.RenderHandler = (*picture)(nil)

func (p *picture) OnRender(element *ui.Element, ctx ui.RenderContext) {
	// ctx.Canvas.UseSolidColor(p.color)
	// ctx.Canvas.DrawRectangle(
	// 	ui.NewPosition(0, 0),
	// 	element.Bounds().Size,
	// )
	ctx.Canvas.DrawImage(p.image,
		ui.NewPosition(0, 0),
		element.Bounds().Size,
	)
}
