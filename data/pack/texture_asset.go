package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
)

type SaveTwoDTextureAssetAction struct {
	locator       AssetLocator
	uri           string
	imageProvider ImageProvider
}

func (a *SaveTwoDTextureAssetAction) Describe() string {
	return fmt.Sprintf("save_twod_texture_asset(uri: %q)", a.uri)
}

func (a *SaveTwoDTextureAssetAction) Run() error {
	image := a.imageProvider.Image()

	textureAsset := &asset.TwoDTexture{
		Width:  uint16(image.Width),
		Height: uint16(image.Height),
		Data:   image.RGBA8Data(),
	}

	out, err := a.locator.Create(a.uri)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := asset.EncodeTwoDTexture(out, textureAsset); err != nil {
		return fmt.Errorf("failed to encode asset: %w", err)
	}
	return nil
}

type SaveCubeTextureAction struct {
	locator       AssetLocator
	uri           string
	imageProvider CubeImageProvider
	format        asset.DataFormat
}

type SaveCubeTextureOption func(a *SaveCubeTextureAction)

func WithFormat(format asset.DataFormat) SaveCubeTextureOption {
	return func(a *SaveCubeTextureAction) {
		a.format = format
	}
}

func (a *SaveCubeTextureAction) Describe() string {
	return fmt.Sprintf("save_cube_texture(uri: %q)", a.uri)
}

func (a *SaveCubeTextureAction) Run() error {
	texture := a.imageProvider.CubeImage()

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

	out, err := a.locator.Create(a.uri)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := asset.EncodeCubeTexture(out, textureAsset); err != nil {
		return fmt.Errorf("failed to encode asset: %w", err)
	}
	return nil
}
