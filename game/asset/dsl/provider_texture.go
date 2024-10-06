package dsl

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/asset/mdl"
)

// Create2DTexture creates a new 2D texture with the specified format and
// source image.
func Create2DTexture(imageProvider Provider[*mdl.Image], opts ...Operation) Provider[*mdl.Texture] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Texture, error) {
			var cfg textureConfig
			for _, opt := range opts {
				if err := opt.Apply(&cfg); err != nil {
					return nil, fmt.Errorf("failed to configure 2D texture: %w", err)
				}
			}

			image, err := imageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get image: %w", err)
			}

			texture := mdl.Create2DTexture(image.Width(), image.Height(), 1, cfg.format.ValueOrDefault(mdl.TextureFormatRGBA8))
			texture.SetName(image.Name())
			texture.SetGenerateMipmaps(cfg.mipmapping)
			texture.SetLayerImage(0, 0, image)
			return texture, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-2d-texture", imageProvider, opts)
		},
	))
}

// CreateCubeTexture creates a new cube texture with the specified format and
// source image.
func CreateCubeTexture(cubeImageProvider Provider[*mdl.CubeImage], opts ...Operation) Provider[*mdl.Texture] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Texture, error) {
			var cfg textureConfig
			for _, opt := range opts {
				if err := opt.Apply(&cfg); err != nil {
					return nil, fmt.Errorf("failed to configure cube texture: %w", err)
				}
			}

			cubeImage, err := cubeImageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get cube image: %w", err)
			}

			frontImage := cubeImage.Side(mdl.CubeSideFront)
			rearImage := cubeImage.Side(mdl.CubeSideRear)
			leftImage := cubeImage.Side(mdl.CubeSideLeft)
			rightImage := cubeImage.Side(mdl.CubeSideRight)
			topImage := cubeImage.Side(mdl.CubeSideTop)
			bottomImage := cubeImage.Side(mdl.CubeSideBottom)

			texture := mdl.CreateCubeTexture(frontImage.Width(), 1, cfg.format.ValueOrDefault(mdl.TextureFormatRGBA16F))
			texture.SetGenerateMipmaps(cfg.mipmapping)
			texture.SetLayerImage(0, 0, frontImage)
			texture.SetLayerImage(0, 1, rearImage)
			texture.SetLayerImage(0, 2, leftImage)
			texture.SetLayerImage(0, 3, rightImage)
			texture.SetLayerImage(0, 4, topImage)
			texture.SetLayerImage(0, 5, bottomImage)
			return texture, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-cube-texture", cubeImageProvider, opts)
		},
	))
}

// CreateCubeMipmapTexture creates a new cube texture with the specified format
// and source mipmap images.
func CreateCubeMipmapTexture(cubeImagesProvider Provider[[]*mdl.CubeImage], opts ...Operation) Provider[*mdl.Texture] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Texture, error) {
			var cfg textureConfig
			for _, opt := range opts {
				if err := opt.Apply(&cfg); err != nil {
					return nil, fmt.Errorf("failed to configure cube texture: %w", err)
				}
			}

			cubeImages, err := cubeImagesProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get cube image: %w", err)
			}

			dimension := cubeImages[0].Side(mdl.CubeSideFront).Width()
			texture := mdl.CreateCubeTexture(dimension, len(cubeImages), cfg.format.ValueOrDefault(mdl.TextureFormatRGBA16F))
			for i, cubeImage := range cubeImages {
				texture.SetLayerImage(i, 0, cubeImage.Side(mdl.CubeSideFront))
				texture.SetLayerImage(i, 1, cubeImage.Side(mdl.CubeSideRear))
				texture.SetLayerImage(i, 2, cubeImage.Side(mdl.CubeSideLeft))
				texture.SetLayerImage(i, 3, cubeImage.Side(mdl.CubeSideRight))
				texture.SetLayerImage(i, 4, cubeImage.Side(mdl.CubeSideTop))
				texture.SetLayerImage(i, 5, cubeImage.Side(mdl.CubeSideBottom))
			}
			return texture, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-cube-mipmap-texture", cubeImagesProvider, opts)
		},
	))
}

type textureConfig struct {
	format     opt.T[mdl.TextureFormat]
	mipmapping bool
}

func (c *textureConfig) SetFormat(format mdl.TextureFormat) {
	c.format = opt.V(format)
}

func (c *textureConfig) Mipmapping() bool {
	return c.mipmapping
}

func (c *textureConfig) SetMipmapping(mipmapping bool) {
	c.mipmapping = mipmapping
}

var defaultCubeTextureProvider = CreateCubeTexture(defaultCubeImageProvider)
