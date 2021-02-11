package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/mokiat/lacking/data/asset"
)

type Window interface {
	SetTitle(title string)
	SetSize(size Size)
	Size() Size
	OpenTemplate(uri string) (*Template, error)
	InstantiateTemplate(template *Template) (Control, error)
	OpenImage(uri string) (Image, error)
	OpenView(viewProvider ViewProvider)
	Destroy()
}

func CreateWindow(driver Driver) (Window, DriverSubscriber) {
	result := &window{
		driver:     driver,
		activeView: CreateView(nil),
	}
	return result, result
}

type window struct {
	driver     Driver
	size       Size
	activeView View
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

func (w *window) OpenTemplate(uri string) (*Template, error) {
	in, err := os.Open(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open template asset: %w", err)
	}
	defer in.Close()

	var templateAsset asset.UITemplate
	if err := asset.DecodeUITemplate(in, &templateAsset); err != nil {
		return nil, fmt.Errorf("failed to decode template: %w", err)
	}
	return buildTemplateFromAsset(templateAsset.Root), nil
}

func (w *window) InstantiateTemplate(template *Template) (Control, error) {
	return Build(BuildContext{
		Window:   w,
		Template: template,
	})
}

func (w *window) OpenImage(uri string) (Image, error) {
	in, err := os.Open(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open image asset: %w", err)
	}
	defer in.Close()

	var imageAsset asset.TwoDTexture
	if err := asset.DecodeTwoDTexture(in, &imageAsset); err != nil {
		return nil, fmt.Errorf("failed to decode image asset: %w", err)
	}
	return w.driver.CreateImage(imageAsset)
}

func (w *window) OpenView(provider ViewProvider) {
	view, err := provider.CreateView(w)
	if err != nil {
		panic(err)
	}
	w.activeView = view
}

func (w *window) Destroy() {
	w.driver.Destroy()
}

func (w *window) OnCreate(d Driver) {
}

func (w *window) OnDestroy(d Driver) {
}

func (w *window) OnResize(d Driver, size Size) {
	w.size = size
	if w.activeView != nil {
		w.activeView.Element().SetBounds(Bounds{
			Position: NewPosition(0, 0),
			Size:     size,
		})
	}
}

func (w *window) OnKeyboardEvent(d Driver, event KeyboardEvent) {
	// log.Printf("keyboard event: %+v", event)
}

func (w *window) OnMouseEvent(d Driver, event MouseEvent) {
	// log.Printf("mouse event: %+v", event)
}

func (w *window) OnRender(d Driver, canvas Canvas) {
	// log.Println("render")

	// canvas.Translate(NewPosition(25, 0))

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
