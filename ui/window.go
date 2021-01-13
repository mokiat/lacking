package ui

import "log"

type Window interface {
	SetTitle(title string)
	SetSize(size Size)
	Size() Size
	// OpenView() // TODO
	Destroy()
}

func CreateWindow(driver Driver) (Window, DriverSubscriber) {
	result := &window{
		driver: driver,
	}
	return result, result
}

type window struct {
	driver     Driver
	size       Size
	activeView View
}

func (w *window) OnCreate(d Driver) {
}

func (w *window) OnDestroy(d Driver) {
}

func (w *window) OnResize(d Driver, size Size) {
	log.Printf("resize: %+v\n", size)
	w.size = size
}

func (w *window) OnKeyboardEvent(d Driver, event KeyboardEvent) {
	log.Printf("keyboard event: %+v", event)
}

func (w *window) OnMouseEvent(d Driver, event MouseEvent) {
	log.Printf("mouse event: %+v", event)
}

func (w *window) OnRender(d Driver, canvas Canvas) {
	log.Println("render")
	if w.activeView != nil {
		renderElement(w.activeView.Element(), RenderContext{
			Canvas: canvas,
			DirtyRegion: Bounds{
				Position: NewPosition(0, 0),
				Size:     w.size,
			},
		})
	}
}

func (w *window) OnCloseRequested(d Driver) {
	log.Println("close requested")
	w.Destroy()
}

func (w *window) SetTitle(title string) {
	w.driver.SetTitle(title)
}

func (w *window) SetSize(size Size) {
	w.driver.SetSize(size)
}

func (w *window) Size() Size {
	return w.driver.Size()
}

func (w *window) Destroy() {
	w.driver.Destroy()
}
