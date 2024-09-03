package dsl

import (
	"fmt"
	"os"

	"github.com/mokiat/gog/opt"

	"github.com/mokiat/lacking/game/asset/mdl"
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

			img, err := mdl.ParseImage(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse image %q: %w", path, err)
			}

			return img, nil
		},

		// digest function
		func() ([]byte, error) {
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %q: %w", path, err)
			}
			return CreateDigest("open-image", path, info.ModTime())
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
			return CreateDigest("resized-image", imageProvider, newWidthProvider, newHeightProvider)
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
			return CreateDigest("cube-image-from-equirectangular", imageProvider)
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
			return CreateDigest("resized-cube-image", imageProvider, newSizeProvider)
		},
	))
}

// IrradianceCubeImage creates an irradiance cube image from the provided
// HDR skybox cube image.
func IrradianceCubeImage(imageProvider Provider[*mdl.CubeImage], opts ...Operation) Provider[*mdl.CubeImage] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.CubeImage, error) {
			var cfg irradianceConfig
			for _, opt := range opts {
				if err := opt.Apply(&cfg); err != nil {
					return nil, fmt.Errorf("failed to configure irradiance cube image: %w", err)
				}
			}

			image, err := imageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting image: %w", err)
			}

			sampleCount := cfg.sampleCount.ValueOrDefault(20)
			return mdl.BuildIrradianceCubeImage(image, sampleCount), nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("irradiance-cube-image", imageProvider, opts)
		},
	))
}

type irradianceConfig struct {
	sampleCount opt.T[int]
}

func (c *irradianceConfig) SetSampleCount(value int) {
	c.sampleCount = opt.V(value)
}

var defaultImageProvider = func() Provider[*mdl.Image] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Image, error) {
			image := mdl.NewImage(1, 1)
			image.SetTexel(0, 0, mdl.Color{
				R: 1.0, G: 0.0, B: 1.0, // purple
			})
			return image, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("default-image")
		},
	))
}()

var defaultCubeImageProvider = func() Provider[*mdl.CubeImage] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.CubeImage, error) {
			image, err := defaultImageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get default image: %w", err)
			}

			result := mdl.NewCubeImage(image.Width())
			result.SetSide(mdl.CubeSideFront, image)
			result.SetSide(mdl.CubeSideRear, image)
			result.SetSide(mdl.CubeSideLeft, image)
			result.SetSide(mdl.CubeSideRight, image)
			result.SetSide(mdl.CubeSideTop, image)
			result.SetSide(mdl.CubeSideBottom, image)
			return result, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("default-cube-image")
		},
	))
}()
