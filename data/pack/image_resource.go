package pack

import (
	"fmt"
	"hash"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sync"

	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/codec/rgbe"
	_ "golang.org/x/image/tiff"
)

func OpenImageResource(uri string) *OpenImageResourceAction {
	return &OpenImageResourceAction{
		uri: uri,
	}
}

var _ ImageProvider = (*OpenImageResourceAction)(nil)

type OpenImageResourceAction struct {
	uri string

	resultMutex  sync.Mutex
	resultDigest []byte
	result       *Image
}

func (a *OpenImageResourceAction) Describe() string {
	return fmt.Sprintf("open_image_resource(uri: %q)", a.uri)
}

func (a *OpenImageResourceAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "open_image_resource", HashableParams{
		"uri": a.uri,
	})
}

func (a *OpenImageResourceAction) Image(ctx *Context) (*Image, error) {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	a.resultMutex.Lock()
	defer a.resultMutex.Unlock()

	digest, err := CalculateDigest(a)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}
	if EqualDigests(digest, a.resultDigest) {
		return a.result, nil
	}

	result, err := a.run(ctx)
	if err != nil {
		return nil, err
	}

	a.result = result
	a.resultDigest = digest
	return result, nil
}

func (a *OpenImageResourceAction) run(ctx *Context) (*Image, error) {
	var img image.Image
	readImage := func(storage Storage) error {
		in, err := storage.OpenResource(a.uri)
		if err != nil {
			return err
		}
		defer in.Close()

		readImg, _, err := image.Decode(in)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
		img = readImg
		return nil
	}
	if err := ctx.IO(readImage); err != nil {
		return nil, err
	}

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
	}, nil
}
