package ui

import (
	"fmt"
	"image"

	"golang.org/x/image/font/opentype"

	"github.com/mokiat/lacking/framework/opengl/ui/internal"
	"github.com/mokiat/lacking/ui"
)

func NewGraphics() *Graphics {
	return &Graphics{
		canvas: internal.NewCanvas(),
	}
}

var _ ui.Graphics = (*Graphics)(nil)

type Graphics struct {
	canvas *internal.Canvas
}

func (g *Graphics) Create() {
	g.canvas.Create()
}

func (g *Graphics) Destroy() {
	g.canvas.Destroy()
}

func (g *Graphics) Begin() {
	g.canvas.Begin()
}

func (g *Graphics) End() {
	g.canvas.End()
}

func (g *Graphics) Resize(size ui.Size) {
	g.canvas.Resize(size.Width, size.Height)
}

func (g *Graphics) ResizeFramebuffer(size ui.Size) {
	g.canvas.ResizeFramebuffer(size.Width, size.Height)
}

func (g *Graphics) CreateImage(img image.Image) (ui.Image, error) {
	result := internal.NewImage()
	result.Allocate(img)
	return result, nil
}

func (g *Graphics) ReleaseImage(resource ui.Image) error {
	image := resource.(*internal.Image)
	image.Release()
	return nil
}

func (g *Graphics) CreateFont(font *opentype.Font) (ui.Font, error) {
	familyName, err := font.Name(nil, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get family name: %w", err)
	}
	subFamilyName, err := font.Name(nil, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-family name: %w", err)
	}

	result := internal.NewFont(familyName, subFamilyName)

	return result, nil // TODO
}

func (g *Graphics) ReleaseFont(resource ui.Font) error {
	// TODO
	return nil
}

func (g *Graphics) Canvas() ui.Canvas {
	return g.canvas
}
