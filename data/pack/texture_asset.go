package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
	gameasset "github.com/mokiat/lacking/game/asset"
)

type SaveTwoDTextureAssetAction struct {
	registry      gameasset.Registry
	id            string
	imageProvider ImageProvider
}

func (a *SaveTwoDTextureAssetAction) Describe() string {
	return fmt.Sprintf("save_twod_texture_asset(id: %q)", a.id)
}

func (a *SaveTwoDTextureAssetAction) Run() error {
	image := a.imageProvider.Image()

	textureAsset := &asset.TwoDTexture{
		Width:     uint16(image.Width),
		Height:    uint16(image.Height),
		WrapModeS: asset.WrapModeRepeat,
		WrapModeT: asset.WrapModeRepeat,
		MagFilter: asset.FilterModeLinear,
		MinFilter: asset.FilterModeLinearMipmapLinear,
		Data:      image.RGBA8Data(),
	}
	if err := a.registry.WriteContent(a.id, textureAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}

type SaveCubeTextureAction struct {
	registry      gameasset.Registry
	id            string
	imageProvider CubeImageProvider
	format        asset.TexelFormat
}

type SaveCubeTextureOption func(a *SaveCubeTextureAction)

func WithFormat(format asset.TexelFormat) SaveCubeTextureOption {
	return func(a *SaveCubeTextureAction) {
		a.format = format
	}
}

func (a *SaveCubeTextureAction) Describe() string {
	return fmt.Sprintf("save_cube_texture(id: %q)", a.id)
}

func (a *SaveCubeTextureAction) Run() error {
	texture := a.imageProvider.CubeImage()

	textureData := func(side CubeSide) []byte {
		switch a.format {
		case asset.TexelFormatRGBA8:
			return texture.RGBA8Data(side)
		case asset.TexelFormatRGBA32F:
			return texture.RGBA32FData(side)
		default:
			panic(fmt.Errorf("unsupported format: %d", a.format))
		}
	}

	textureAsset := &asset.CubeTexture{
		Dimension: uint16(texture.Dimension),
		MagFilter: asset.FilterModeLinear,
		MinFilter: asset.FilterModeLinear,
		Format:    a.format,
	}
	textureAsset.FrontSide = asset.CubeTextureSide{
		Data: textureData(CubeSideFront),
	}
	textureAsset.BackSide = asset.CubeTextureSide{
		Data: textureData(CubeSideRear),
	}
	textureAsset.LeftSide = asset.CubeTextureSide{
		Data: textureData(CubeSideLeft),
	}
	textureAsset.RightSide = asset.CubeTextureSide{
		Data: textureData(CubeSideRight),
	}
	textureAsset.TopSide = asset.CubeTextureSide{
		Data: textureData(CubeSideTop),
	}
	textureAsset.BottomSide = asset.CubeTextureSide{
		Data: textureData(CubeSideBottom),
	}

	if err := a.registry.WriteContent(a.id, textureAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}
