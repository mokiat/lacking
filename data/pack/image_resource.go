package pack

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/codec/rgbe"
	"github.com/mokiat/goexr/exr"
	"github.com/mokiat/lacking/util/resource"
	_ "golang.org/x/image/tiff"
)

type OpenImageResourceAction struct {
	locator resource.ReadLocator
	uri     string
	image   *Image
}

func (a *OpenImageResourceAction) Describe() string {
	return fmt.Sprintf("open_image_resource(%q)", a.uri)
}

func (a *OpenImageResourceAction) Image() *Image {
	if a.image == nil {
		panic("reading data from unprocessed action")
	}
	return a.image
}

func (a *OpenImageResourceAction) Run() error {
	in, err := a.locator.ReadResource(a.uri)
	if err != nil {
		return fmt.Errorf("failed to open image resource: %w", err)
	}
	defer in.Close()

	img, err := ParseImageResource(in)
	if err != nil {
		return err
	}

	a.image = img
	return nil
}

func ParseImageResource(in io.Reader) (*Image, error) {
	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return BuildImageResource(img), nil
}

func BuildImageResource(img image.Image) *Image {
	imgStartX := img.Bounds().Min.X
	imgStartY := img.Bounds().Min.Y
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	texels := make([][]Color, height)
	for y := 0; y < height; y++ {
		texels[y] = make([]Color, width)
		for x := 0; x < width; x++ {
			switch img := img.(type) {
			case hdr.Image:
				r, g, b, a := img.HDRAt(imgStartX+x, imgStartY+y).HDRPixel()
				texels[y][x] = Color{
					R: r,
					G: g,
					B: b,
					A: a,
				}
			case *exr.RGBAImage:
				clr := img.At(x, y).(exr.RGBAColor)
				texels[y][x] = Color{
					R: float64(clr.R),
					G: float64(clr.G),
					B: float64(clr.B),
					A: float64(clr.A),
				}
			default:
				r, g, b, a := img.At(imgStartX+x, imgStartY+y).RGBA()
				texels[y][x] = Color{
					R: float64(float64((r>>8)&0xFF) / 255.0),
					G: float64(float64((g>>8)&0xFF) / 255.0),
					B: float64(float64((b>>8)&0xFF) / 255.0),
					A: float64(float64((a>>8)&0xFF) / 255.0),
				}
			}
		}
	}
	return &Image{
		Width:  width,
		Height: height,
		Texels: texels,
	}
}
