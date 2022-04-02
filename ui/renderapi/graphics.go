package renderapi

import (
	"image"

	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/ui/renderapi/internal"
	"github.com/mokiat/lacking/ui/renderapi/plugin"
	"golang.org/x/image/font/opentype"
)

func NewGraphics(api render.API, shaders plugin.ShaderCollection) *Graphics {
	renderer := internal.NewRenderer(api, shaders)
	return &Graphics{
		api:         api,
		renderer:    renderer,
		fontFactory: internal.NewFontFactory(api, renderer),
		canvas:      internal.NewCanvas(api, renderer),
	}
}

var _ ui.Graphics = (*Graphics)(nil)

type Graphics struct {
	api         render.API
	renderer    *internal.Renderer
	fontFactory *internal.FontFactory
	canvas      *internal.Canvas
}

func (g *Graphics) Create() {
	g.renderer.Init()
	g.fontFactory.Init()
}

func (g *Graphics) Destroy() {
	defer g.renderer.Free()
	defer g.fontFactory.Free()
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
	bounds := img.Bounds()
	size := ui.NewSize(bounds.Dx(), bounds.Dy())

	texture := g.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           size.Width,
		Height:          size.Height,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeLinear,
		Mipmapping:      false,
		GammaCorrection: true, // TODO: Rethink this. Do we need it for UI where alpha is either way linear?
		Format:          render.DataFormatRGBA8,
		Data:            internal.ImgToRGBA8(img),
	})
	return internal.NewImage(texture, size), nil
}

func (g *Graphics) ReleaseImage(resource ui.Image) error {
	image := resource.(*internal.Image)
	image.Destroy()
	return nil
}

func (g *Graphics) CreateFont(font *opentype.Font) (ui.Font, error) {
	result := g.fontFactory.CreateFont(font)
	return result, nil
}

func (g *Graphics) ReleaseFont(resource ui.Font) error {
	font := resource.(*internal.Font)
	font.Destroy()
	return nil
}

func (g *Graphics) Canvas() ui.Canvas {
	return g.canvas
}
