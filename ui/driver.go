package ui

type DriverSubscriber interface {
	OnCreate(d Driver)
	OnDestroy(d Driver)
	OnContentResize(d Driver, size Size)
	OnKeyboardEvent(d Driver, event KeyboardEvent)
	OnRender(d Driver, canvas Canvas)
	OnCloseRequested(d Driver)
}

type Driver interface {
	SetTitle(title string)
	SetSize(size Size)
	Size() Size
	Redraw()
	Destroy()
}
