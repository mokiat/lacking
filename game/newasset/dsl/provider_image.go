package dsl

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/codec/rgbe"
	"github.com/mokiat/goexr/exr"
	_ "golang.org/x/image/tiff"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

// OpenImage opens an image file from the provided path.
func OpenImage(path string) Provider[*mdl.Image] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Image, error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open image file %q: %w", path, err)
			}
			defer file.Close()

			img, err := parseGoImage(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse image %q: %w", path, err)
			}

			return buildImageResource(img), nil
		},

		// digest function
		func() ([]byte, error) {
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %q: %w", path, err)
			}
			return digestItems("open-image", path, info.ModTime())
		},
	))
}

// ResizedImage returns an image with the provided dimensions.
func ResizedImage(imageProvider Provider[*mdl.Image], newWidthProvider, newHeightProvider Provider[int]) Provider[*mdl.Image] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Image, error) {
			newWidth, err := newWidthProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting new width: %w", err)
			}

			newHeight, err := newHeightProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting new height: %w", err)
			}

			image, err := imageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting image: %w", err)
			}
			return image.Scale(newWidth, newHeight), nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("resized-image", imageProvider, newWidthProvider, newHeightProvider)
		},
	))
}

// CubeImageFromEquirectangular creates a cube image from an
// equirectangular image.
func CubeImageFromEquirectangular(imageProvider Provider[*mdl.Image]) Provider[*mdl.CubeImage] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.CubeImage, error) {
			image, err := imageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting image: %w", err)
			}

			frontImage := mdl.BuildCubeSideFromEquirectangular(mdl.CubeSideFront, image)
			rearImage := mdl.BuildCubeSideFromEquirectangular(mdl.CubeSideRear, image)
			leftImage := mdl.BuildCubeSideFromEquirectangular(mdl.CubeSideLeft, image)
			rightImage := mdl.BuildCubeSideFromEquirectangular(mdl.CubeSideRight, image)
			topImage := mdl.BuildCubeSideFromEquirectangular(mdl.CubeSideTop, image)
			bottomImage := mdl.BuildCubeSideFromEquirectangular(mdl.CubeSideBottom, image)

			dstImage := mdl.NewCubeImage(frontImage.Width())
			dstImage.SetSide(mdl.CubeSideFront, frontImage)
			dstImage.SetSide(mdl.CubeSideRear, rearImage)
			dstImage.SetSide(mdl.CubeSideLeft, leftImage)
			dstImage.SetSide(mdl.CubeSideRight, rightImage)
			dstImage.SetSide(mdl.CubeSideTop, topImage)
			dstImage.SetSide(mdl.CubeSideBottom, bottomImage)
			return dstImage, nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("cube-image-from-equirectangular", imageProvider)
		},
	))
}

// ResizedCubeImage returns a cube image with the provided dimensions.
func ResizedCubeImage(imageProvider Provider[*mdl.CubeImage], newSizeProvider Provider[int]) Provider[*mdl.CubeImage] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.CubeImage, error) {
			newSize, err := newSizeProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting new size: %w", err)
			}

			image, err := imageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting image: %w", err)
			}
			return image.Scale(newSize), nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("resized-cube-image", imageProvider, newSizeProvider)
		},
	))
}

// IrradianceCubeImage creates an irradiance cube image from the provided
// HDR skybox cube image.
func IrradianceCubeImage(imageProvider Provider[*mdl.CubeImage], opts ...Operation) Provider[*mdl.CubeImage] {
	cfg := irradianceConfig{
		sampleCount: 20,
	}
	for _, opt := range opts {
		opt.Apply(&cfg)
	}

	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.CubeImage, error) {
			image, err := imageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting image: %w", err)
			}

			return mdl.BuildIrradianceCubeImage(image, cfg.sampleCount), nil
		},

		// digest function
		func() ([]byte, error) {
			return digestItems("irradiance-cube-image", imageProvider, cfg.sampleCount)
		},
	))
}

func parseGoImage(in io.Reader) (image.Image, error) {
	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

func buildImageResource(img image.Image) *mdl.Image {
	imgStartX := img.Bounds().Min.X
	imgStartY := img.Bounds().Min.Y
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	image := mdl.NewImage(width, height)
	for y := range height {
		for x := range width {
			atX := imgStartX + x
			atY := imgStartY + y
			switch img := img.(type) {
			case hdr.Image:
				r, g, b, a := img.HDRAt(atX, atY).HDRPixel()
				image.SetTexel(x, y, mdl.Color{
					R: r,
					G: g,
					B: b,
					A: a,
				})
			case *exr.RGBAImage:
				clr := img.At(atX, atY).(exr.RGBAColor)
				image.SetTexel(x, y, mdl.Color{
					R: float64(clr.R),
					G: float64(clr.G),
					B: float64(clr.B),
					A: float64(clr.A),
				})
			default:
				c := color.NRGBAModel.Convert(img.At(atX, atY)).(color.NRGBA)
				image.SetTexel(x, y, mdl.Color{
					R: float64(float64(c.R) / 255.0),
					G: float64(float64(c.G) / 255.0),
					B: float64(float64(c.B) / 255.0),
					A: float64(float64(c.A) / 255.0),
				})
			}
		}
	}
	return image
}

type irradianceConfig struct {
	sampleCount int
}

func (c *irradianceConfig) SampleCount() int {
	return c.sampleCount
}

func (c *irradianceConfig) SetSampleCount(value int) {
	c.sampleCount = value
}
