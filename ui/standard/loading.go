package standard

import (
	"fmt"
	"time"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
)

func init() {
	ui.RegisterControlBuilder("Loading", ui.ControlBuilderFunc(func(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (ui.Control, error) {
		return BuildLoading(ctx, template, layoutConfig)
	}))
}

// Loading represents a Control that indicates a loading
// operation that does not have a deterministic progress.
type Loading interface {
	ui.Control
}

// BuildLoading constructs a new Loading control.
func BuildLoading(ctx *ui.Context, template *ui.Template, layoutConfig ui.LayoutConfig) (Loading, error) {
	result := &loading{
		element:    ctx.CreateElement(),
		lastUpdate: time.Now(),
	}
	result.element.SetLayoutConfig(layoutConfig)
	result.element.SetEssence(result)
	if err := result.ApplyAttributes(template.Attributes()); err != nil {
		return nil, err
	}
	return result, nil
}

var _ ui.ElementRenderHandler = (*loading)(nil)

type loading struct {
	element *ui.Element

	image      ui.Image
	angle      sprec.Angle
	lastUpdate time.Time
}

func (l *loading) Element() *ui.Element {
	return l.element
}

func (l *loading) ApplyAttributes(attributes ui.AttributeSet) error {
	if err := l.element.ApplyAttributes(attributes); err != nil {
		return err
	}
	if src, ok := attributes.StringAttribute("src"); ok {
		context := l.element.Context()
		img, err := context.OpenImage(src)
		if err != nil {
			return fmt.Errorf("failed to open image: %w", err)
		}
		l.image = img
	}
	return nil
}

func (l *loading) OnRender(element *ui.Element, canvas ui.Canvas) {
	currentTime := time.Now()
	l.angle += sprec.Degrees(360 * float32(currentTime.Sub(l.lastUpdate).Seconds()))
	l.lastUpdate = currentTime

	cs := (sprec.Cos(l.angle) + 1.0) / 2.0
	canvas.SetSolidColor(ui.RGB(0, 128, 255))
	canvas.FillRectangle(
		ui.NewPosition(0, 0),
		ui.NewSize(
			int(float32(element.Bounds().Width)*cs),
			int(float32(element.Bounds().Height)*cs),
		),
	)
	// Force redraw
	context := l.element.Context()
	context.Window().Invalidate()
}
