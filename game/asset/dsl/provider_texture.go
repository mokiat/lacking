package dsl

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/asset/mdl"
)

// Create2DTexture creates a new 2D texture with the specified format and
// source image.
func Create2DTexture(imageProvider Provider[*mdl.Image], opts ...Operation) Provider[*mdl.Texture] {
	var cfg textureConfig
	for _, opt := range opts {
		opt.Apply(&cfg)
	}

	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Texture, error) {
			image, err := imageProvider.Get()
			if err != nil {
				return nil, fmt.Errorf("failed to get image: %w", err)
			}

			var texture mdl.Texture
			texture.SetName(image.Name())
			texture.SetKind(mdl.TextureKind2D)
			texture.SetFormat(cfg.format.ValueOrDefault(mdl.TextureFormatRGBA8))
			texture.SetGenerateMipmaps(true)
			texture.Resize(image.Width(), image.Height())
			texture.SetLayerImage(0, image)
			return &texture, nil
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
	var cfg textureConfig
	for _, opt := range opts {
		opt.Apply(&cfg)
	}

	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Texture, error) {
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

			var texture mdl.Texture
			texture.SetKind(mdl.TextureKindCube)
			texture.SetFormat(cfg.format.ValueOrDefault(mdl.TextureFormatRGBA16F))
			texture.Resize(frontImage.Width(), frontImage.Height())
			texture.SetLayerImage(0, frontImage)
			texture.SetLayerImage(1, rearImage)
			texture.SetLayerImage(2, leftImage)
			texture.SetLayerImage(3, rightImage)
			texture.SetLayerImage(4, topImage)
			texture.SetLayerImage(5, bottomImage)
			return &texture, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-cube-texture", cubeImageProvider, opts)
		},
	))
}

type textureConfig struct {
	format opt.T[mdl.TextureFormat]
}

func (c *textureConfig) SetFormat(format mdl.TextureFormat) {
	c.format = opt.V(format)
}

var defaultCubeTextureProvider = CreateCubeTexture(defaultCubeImageProvider)
