package std

import (
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
)

// DropZoneCallbackData can be used to specify callback data for a DropZone
// component.
type DropZoneCallbackData struct {
	OnDrop func(paths []string) bool
}

var dropZoneDefaultCallbackData = DropZoneCallbackData{
	OnDrop: func([]string) bool { return false },
}

// DropZone is a transparent component that handles file drop events.
//
// It is intended to be used as a container for a different component that
// provides a visual aid (e.g. an upload icon or some type of viewport).
var DropZone = co.Define(&dropZoneComponent{})

type dropZoneComponent struct {
	Properties co.Properties `co:"properties"`

	onDrop func(paths []string) bool
}

func (c *dropZoneComponent) OnUpsert() {
	callbackData := co.GetOptionalCallbackData(c.Properties, dropZoneDefaultCallbackData)
	c.onDrop = callbackData.OnDrop
}

func (c *dropZoneComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	if event.Type != ui.MouseEventTypeDrop {
		return false
	}
	filePayload, ok := event.Payload.(ui.FilepathPayload)
	if !ok {
		return false
	}
	return c.onDrop(filePayload.Paths)
}

func (c *dropZoneComponent) Render() co.Instance {
	return co.New(co.Element, func() {
		co.WithLayoutData(c.Properties.LayoutData())
		co.WithData(co.ElementData{
			Essence: c,
			Layout:  layout.Fill(),
		})
		co.WithChildren(c.Properties.Children())
	})
}
