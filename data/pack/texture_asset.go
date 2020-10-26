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
}

func (a *SaveCubeTextureAction) Describe() string {
	return fmt.Sprintf("save_cube_texture(uri: %q)", a.uri)
}

func (a *SaveCubeTextureAction) Run() error {
	texture := a.imageProvider.CubeImage()

	textureAsset := &asset.CubeTexture{
		Dimension: uint16(texture.Dimension),
	}
	textureAsset.Sides[asset.TextureSideFront] = asset.CubeTextureSide{
		Data: texture.RGBA8Data(CubeSideFront),
	}
	textureAsset.Sides[asset.TextureSideBack] = asset.CubeTextureSide{
		Data: texture.RGBA8Data(CubeSideRear),
	}
	textureAsset.Sides[asset.TextureSideLeft] = asset.CubeTextureSide{
		Data: texture.RGBA8Data(CubeSideLeft),
	}
	textureAsset.Sides[asset.TextureSideRight] = asset.CubeTextureSide{
		Data: texture.RGBA8Data(CubeSideRight),
	}
	textureAsset.Sides[asset.TextureSideTop] = asset.CubeTextureSide{
		Data: texture.RGBA8Data(CubeSideTop),
	}
	textureAsset.Sides[asset.TextureSideBottom] = asset.CubeTextureSide{
		Data: texture.RGBA8Data(CubeSideBottom),
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
