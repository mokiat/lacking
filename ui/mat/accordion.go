package mat

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/util/optional"
)

var (
	AccordionHeaderPadding        = 2
	AccordionHeaderContentSpacing = 5
	AccordionHeaderIconSize       = 24
	AccordionHeaderFontSize       = float32(20)
	AccordionHeaderFontFile       = "mat:///roboto-regular.ttf"
	AccordionExpandedIconFile     = "mat:///expanded.png"
	AccordionCollapsedIconFile    = "mat:///collapsed.png"
)

// AccordionData holds the data for an Accordion container.
type AccordionData struct {
	Title    string
	Expanded bool
}

var defaultAccordionData = AccordionData{}

// AccordionCallbackData holds the callback data for an Accordion container.
type AccordionCallbackData struct {
	OnToggle func()
}

var defaultAccordionCallbackData = AccordionCallbackData{
	OnToggle: func() {},
}

// Accordion is a container that hides a big chunk of UI until it is expanded.
var Accordion = co.Define(func(props co.Properties, scope co.Scope) co.Instance {
	var (
		data         = co.GetOptionalData(props, defaultAccordionData)
		callbackData = co.GetOptionalCallbackData(props, defaultAccordionCallbackData)
	)

	headerEssence := co.UseState(func() *accordionHeaderEssence {
		return &accordionHeaderEssence{
			ButtonBaseEssence: NewButtonBaseEssence(callbackData.OnToggle),
		}
	}).Get()

	var icon *ui.Image
	if data.Expanded {
		icon = co.OpenImage(scope, AccordionExpandedIconFile)
	} else {
		icon = co.OpenImage(scope, AccordionCollapsedIconFile)
	}

	return co.New(Element, func() {
		co.WithData(ElementData{
			Layout: NewVerticalLayout(VerticalLayoutSettings{
				ContentAlignment: AlignmentLeft,
			}),
		})
		co.WithLayoutData(props.LayoutData())

		co.WithChild("header", co.New(Element, func() {
			co.WithData(ElementData{
				Essence: headerEssence,
				Padding: ui.Spacing{
					Left:   AccordionHeaderPadding,
					Right:  AccordionHeaderPadding,
					Top:    AccordionHeaderPadding,
					Bottom: AccordionHeaderPadding,
				},
				Layout: NewHorizontalLayout(HorizontalLayoutSettings{
					ContentAlignment: AlignmentCenter,
					ContentSpacing:   AccordionHeaderContentSpacing,
				}),
			})
			co.WithLayoutData(LayoutData{
				GrowHorizontally: true,
			})

			co.WithChild("icon", co.New(Picture, func() {
				co.WithData(PictureData{
					Image:      icon,
					ImageColor: optional.Value(OnPrimaryLightColor),
					Mode:       ImageModeFit,
				})
				co.WithLayoutData(LayoutData{
					Width:  optional.Value(AccordionHeaderIconSize),
					Height: optional.Value(AccordionHeaderIconSize),
				})
			}))

			co.WithChild("title", co.New(Label, func() {
				co.WithData(LabelData{
					Font:      co.OpenFont(scope, AccordionHeaderFontFile),
					FontSize:  optional.Value(AccordionHeaderFontSize),
					FontColor: optional.Value(OnPrimaryLightColor),
					Text:      data.Title,
				})
			}))
		}))

		if data.Expanded {
			for _, child := range props.Children() {
				co.WithChild(child.Key(), child)
			}
		}
	})
})

var _ ui.ElementRenderHandler = (*accordionHeaderEssence)(nil)

type accordionHeaderEssence struct {
	*ButtonBaseEssence
}

func (e *accordionHeaderEssence) OnRender(element *ui.Element, canvas *ui.Canvas) {
	var backgroundColor ui.Color
	switch e.State() {
	case ButtonStateOver:
		backgroundColor = PrimaryLightColor.Overlay(HoverOverlayColor)
	case ButtonStateDown:
		backgroundColor = PrimaryLightColor.Overlay(PressOverlayColor)
	default:
		backgroundColor = PrimaryLightColor
	}

	size := element.Bounds().Size
	width := float32(size.Width)
	height := float32(size.Height)

	canvas.Reset()
	canvas.SetStrokeSize(1.0)
	canvas.SetStrokeColor(OutlineColor)
	canvas.Rectangle(
		sprec.ZeroVec2(),
		sprec.NewVec2(width, height),
	)
	if !backgroundColor.Transparent() {
		canvas.Fill(ui.Fill{
			Color: backgroundColor,
		})
	}
	canvas.Stroke()
}
