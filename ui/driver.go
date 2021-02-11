package ui

import "github.com/mokiat/lacking/data/asset"

type DriverSubscriber interface {
	OnCreate(d Driver)
	OnDestroy(d Driver)
	OnResize(d Driver, size Size)
	OnKeyboardEvent(d Driver, event KeyboardEvent)
	OnMouseEvent(d Driver, event MouseEvent)
	OnRender(d Driver, canvas Canvas)
	OnCloseRequested(d Driver)
}

type Driver interface {
	SetTitle(title string)
	SetSize(size Size)
	Size() Size
	Redraw()
	Destroy()

	CreateImage(data asset.TwoDTexture) (Image, error)
}
