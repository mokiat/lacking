package pack

import (
	"fmt"
	"image"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/nfnt/resize"
)

type ImageProvider interface {
	Image() (Image, error)
}

type ImageResourceFile struct {
	Resource
}

func (f *ImageResourceFile) Image() (Image, error) {
	file, err := os.Open(f.filename)
	if err != nil {
		return Image{}, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return Image{}, fmt.Errorf("failed to decode image: %w", err)
	}
	return Image{img: img}, nil
}

type Color struct {
	R byte
	G byte
	B byte
	A byte
}

type Image struct {
	img image.Image
}

func (i Image) Width() int {
	return i.img.Bounds().Dx()
}

func (i Image) Height() int {
	return i.img.Bounds().Dy()
}

func (i Image) IsSquare() bool {
	return i.Width() == i.Height()
}

func (i Image) Scale(newWidth, newHeight int) {
	i.img = resize.Resize(uint(newWidth), uint(newHeight), i.img, resize.Bicubic)
}

func (i Image) Texel(x, y int) Color {
	texel := i.img.At(i.img.Bounds().Min.X+x, i.img.Bounds().Min.Y+y)
	r, g, b, a := texel.RGBA()
	return Color{
		R: byte(r >> 8),
		G: byte(g >> 8),
		B: byte(b >> 8),
		A: byte(a >> 8),
	}
}

func (i Image) RGBAData() []byte {
	width := i.Width()
	height := i.Height()
	data := make([]byte, 4*width*height)
	offset := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			texel := i.Texel(x, height-y-1)
			data[offset+0] = texel.R
			data[offset+1] = texel.G
			data[offset+2] = texel.B
			data[offset+3] = texel.A
			offset += 4
		}
	}
	return data
}
