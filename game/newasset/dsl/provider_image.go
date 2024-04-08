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
	"github.com/mokiat/gomath/dprec"
	_ "golang.org/x/image/tiff"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func OpenImage(path string) Provider[*mdl.Image] {
	get := func() (*mdl.Image, error) {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open image file: %w", err)
		}
		defer file.Close()

		img, err := parseGoImage(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image: %w", err)
		}

		return buildImageResource(img), nil
	}

	digest := func() ([]byte, error) {
		return digestItems("open-image", path)
	}

	return OnceProvider(FuncProvider(get, digest))
}

func CubeImageFromEquirectangular(imageProvider Provider[*mdl.Image]) Provider[*mdl.CubeImage] {
	get := func() (*mdl.CubeImage, error) {
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
	}

	digest := func() ([]byte, error) {
		return digestItems("cube-image-from-equirectangular", imageProvider)
	}

	return OnceProvider(FuncProvider(get, digest))
}

func ResizedCubeImage(imageProvider Provider[*mdl.CubeImage], newSize int) Provider[*mdl.CubeImage] {
	get := func() (*mdl.CubeImage, error) {
		image, err := imageProvider.Get()
		if err != nil {
			return nil, fmt.Errorf("error getting image: %w", err)
		}
		return image.Scale(newSize), nil
	}

	digest := func() ([]byte, error) {
		return digestItems("resized-cube-image", imageProvider, newSize)
	}

	return OnceProvider(FuncProvider(get, digest))
}

func IrradianceCubeImage(imageProvider Provider[*mdl.CubeImage]) Provider[*mdl.CubeImage] {
	get := func() (*mdl.CubeImage, error) {
		image, err := imageProvider.Get()
		if err != nil {
			return nil, fmt.Errorf("error getting image: %w", err)
		}

		return mdl.BuildIrradianceCubeImage(image, 10), nil
	}

	digest := func() ([]byte, error) {
		return digestItems("irradiance-cube-image", imageProvider)
	}

	return OnceProvider(FuncProvider(get, digest))
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
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			switch img := img.(type) {
			case hdr.Image:
				r, g, b, a := img.HDRAt(imgStartX+x, imgStartY+y).HDRPixel()
				image.SetTexel(x, y, mdl.Color{
					R: r,
					G: g,
					B: b,
					A: a,
				})
			case *exr.RGBAImage:
				clr := img.At(x, y).(exr.RGBAColor)
				image.SetTexel(x, y, mdl.Color{
					R: float64(clr.R),
					G: float64(clr.G),
					B: float64(clr.B),
					A: float64(clr.A),
				})
			default:
				c := color.NRGBAModel.Convert(img.At(imgStartX+x, imgStartY+y)).(color.NRGBA)
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

func BuildCubeSideFromEquirectangular(side mdl.CubeSide, imageProvider Provider[*mdl.Image]) Provider[*mdl.Image] {
	get := func() (*mdl.Image, error) {
		srcImage, err := imageProvider.Get()
		if err != nil {
			return nil, fmt.Errorf("error getting image: %w", err)
		}

		dimension := srcImage.Height() / 2
		dstImage := mdl.NewImage(dimension, dimension)

		uv := dprec.ZeroVec2()
		startU := 0.0
		deltaU := 1.0 / float64(dimension-1)
		startV := 1.0
		deltaV := -1.0 / float64(dimension-1)

		uv.Y = startV
		for y := 0; y < dimension; y++ {
			uv.X = startU
			for x := 0; x < dimension; x++ {
				dstImage.SetTexel(x, y, srcImage.TexelUV(mdl.UVWToEquirectangularUV(mdl.CubeUVToUVW(side, uv))))
				uv.X += deltaU
			}
			uv.Y += deltaV
		}
		return dstImage, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("build-cube-side-from-equirectangular", uint8(side), imageProvider)
	}

	return OnceProvider(FuncProvider(get, digest))
}
