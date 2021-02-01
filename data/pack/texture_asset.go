package pack

import (
	"fmt"
	"hash"

	"github.com/mokiat/lacking/data/asset"
)

func SaveTwoDTextureAsset(uri string, imageProvider ImageProvider) *SaveTwoDTextureAssetAction {
	return &SaveTwoDTextureAssetAction{
		uri:           uri,
		imageProvider: imageProvider,
	}
}

var _ Action = (*SaveTwoDTextureAssetAction)(nil)

type SaveTwoDTextureAssetAction struct {
	uri           string
	imageProvider ImageProvider
}

func (a *SaveTwoDTextureAssetAction) Describe() string {
	return fmt.Sprintf("save_twod_texture_asset(uri: %q)", a.uri)
}

func (a *SaveTwoDTextureAssetAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "save_twod_texture_asset", HashableParams{
		"uri":   a.uri,
		"image": a.imageProvider,
	})
}

func (a *SaveTwoDTextureAssetAction) Run(ctx *Context) error {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	image, err := a.imageProvider.Image(ctx)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}
	textureAsset := &asset.TwoDTexture{
		Width:  uint16(image.Width),
		Height: uint16(image.Height),
		Data:   image.RGBA8Data(),
	}

	return ctx.IO(func(storage Storage) error {
		out, err := storage.CreateAsset(a.uri)
		if err != nil {
			return err
		}
		defer out.Close()

		if err := asset.EncodeTwoDTexture(out, textureAsset); err != nil {
			return fmt.Errorf("failed to encode asset: %w", err)
		}
		return nil
	})
}

func SaveCubeTextureAsset(uri string, imageProvider CubeImageProvider) *SaveCubeTextureAssetAction {
	return &SaveCubeTextureAssetAction{
		uri:           uri,
		imageProvider: imageProvider,
		format:        asset.DataFormatRGBA8,
	}
}

var _ Action = (*SaveCubeTextureAssetAction)(nil)

type SaveCubeTextureAssetAction struct {
	uri           string
	imageProvider CubeImageProvider
	format        asset.DataFormat
}

func (a *SaveCubeTextureAssetAction) WithFormat(format asset.DataFormat) *SaveCubeTextureAssetAction {
	a.format = format
	return a
}

func (a *SaveCubeTextureAssetAction) Describe() string {
	return fmt.Sprintf("save_cube_texture_asset(uri: %q)", a.uri)
}

func (a *SaveCubeTextureAssetAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "save_cube_texture_asset", HashableParams{
		"uri":    a.uri,
		"image":  a.imageProvider,
		"format": int(a.format),
	})
}

func (a *SaveCubeTextureAssetAction) Run(ctx *Context) error {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	texture, err := a.imageProvider.CubeImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cube image: %w", err)
	}

	textureData := func(side CubeSide) []byte {
		switch a.format {
		case asset.DataFormatRGBA8:
			return texture.RGBA8Data(side)
		case asset.DataFormatRGBA32F:
			return texture.RGBA32FData(side)
		default:
			panic(fmt.Errorf("unknown format: %d", a.format))
		}
	}

	textureAsset := &asset.CubeTexture{
		Dimension: uint16(texture.Dimension),
		Format:    a.format,
	}
	textureAsset.Sides[asset.TextureSideFront] = asset.CubeTextureSide{
		Data: textureData(CubeSideFront),
	}
	textureAsset.Sides[asset.TextureSideBack] = asset.CubeTextureSide{
		Data: textureData(CubeSideRear),
	}
	textureAsset.Sides[asset.TextureSideLeft] = asset.CubeTextureSide{
		Data: textureData(CubeSideLeft),
	}
	textureAsset.Sides[asset.TextureSideRight] = asset.CubeTextureSide{
		Data: textureData(CubeSideRight),
	}
	textureAsset.Sides[asset.TextureSideTop] = asset.CubeTextureSide{
		Data: textureData(CubeSideTop),
	}
	textureAsset.Sides[asset.TextureSideBottom] = asset.CubeTextureSide{
		Data: textureData(CubeSideBottom),
	}

	return ctx.IO(func(storage Storage) error {
		out, err := storage.CreateAsset(a.uri)
		if err != nil {
			return err
		}
		defer out.Close()

		if err := asset.EncodeCubeTexture(out, textureAsset); err != nil {
			return fmt.Errorf("failed to encode asset: %w", err)
		}
		return nil
	})
}
