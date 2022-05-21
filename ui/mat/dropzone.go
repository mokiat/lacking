package mat

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
)

// DropZoneCallbackData can be used to specify callback data for a DropZone
// component.
type DropZoneCallbackData struct {
	OnDrop func(paths []string) bool
}

var defaultDropZoneCallbackData = DropZoneCallbackData{
	OnDrop: func(paths []string) bool {
		return false
	},
}

// DropZone is a component that handles file drop events.
// It is intended to be used as a container for a different component that
// provides a visual aid (e.g. an upload icon or some type of viewport).
var DropZone = co.Define(func(props co.Properties) co.Instance {
	var (
		callbackData = co.GetOptionalCallbackData(props, defaultDropZoneCallbackData)
	)

	essence := co.UseState(func() *dropZoneEssence {
		return &dropZoneEssence{}
	}).Get()
	essence.onDrop = callbackData.OnDrop

	return co.New(Element, func() {
		co.WithData(ElementData{
			Essence: essence,
			Layout:  NewFillLayout(),
		})
		co.WithLayoutData(props.LayoutData())
		co.WithChildren(props.Children())
	})
})

var _ ui.ElementMouseHandler = (*dropZoneEssence)(nil)

type dropZoneEssence struct {
	onDrop func(paths []string) bool
}

func (e *dropZoneEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	if event.Type != ui.MouseEventTypeDrop {
		return false
	}

	filePayload, ok := event.Payload.(ui.FilepathPayload)
	if !ok {
		return false
	}
	return e.onDrop(filePayload.Paths)
}
