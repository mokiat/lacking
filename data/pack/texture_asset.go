package pack

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
)

func BuildTwoDTextureAsset(image *Image) *asset.TwoDTexture {
	return &asset.TwoDTexture{
		Width:  uint16(image.Width),
		Height: uint16(image.Height),
		Flags:  asset.TextureFlagMipmapping,
		Format: asset.TexelFormatRGBA8,
		Data:   image.RGBA8Data(),
	}
}

type SaveTwoDTextureAssetAction struct {
	resource      asset.Resource
	imageProvider ImageProvider
}

func (a *SaveTwoDTextureAssetAction) Describe() string {
	return fmt.Sprintf("save_twod_texture_asset(%q)", a.resource.Name())
}

func (a *SaveTwoDTextureAssetAction) Run() error {
	image := a.imageProvider.Image()
	textureAsset := BuildTwoDTextureAsset(image)
	if err := a.resource.WriteContent(textureAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}

func BuildCubeTextureAsset(image *CubeImage, format asset.TexelFormat) *asset.CubeTexture {
	textureData := func(side CubeSide) []byte {
		switch format {
		case asset.TexelFormatRGBA8:
			return image.RGBA8Data(side)
		case asset.TexelFormatRGBA16F:
			return image.RGBA16FData(side)
		case asset.TexelFormatRGBA32F:
			return image.RGBA32FData(side)
		default:
			panic(fmt.Errorf("unsupported format: %d", format))
		}
	}
	return &asset.CubeTexture{
		Dimension: uint16(image.Dimension),
		Flags:     asset.TextureFlagNone,
		Format:    format,
		FrontSide: asset.CubeTextureSide{
			Data: textureData(CubeSideFront),
		},
		BackSide: asset.CubeTextureSide{
			Data: textureData(CubeSideRear),
		},
		LeftSide: asset.CubeTextureSide{
			Data: textureData(CubeSideLeft),
		},
		RightSide: asset.CubeTextureSide{
			Data: textureData(CubeSideRight),
		},
		TopSide: asset.CubeTextureSide{
			Data: textureData(CubeSideTop),
		},
		BottomSide: asset.CubeTextureSide{
			Data: textureData(CubeSideBottom),
		},
	}
}

type SaveCubeTextureAction struct {
	resource      asset.Resource
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
	return fmt.Sprintf("save_cube_texture(%q)", a.resource.Name())
}

func (a *SaveCubeTextureAction) Run() error {
	image := a.imageProvider.CubeImage()
	textureAsset := BuildCubeTextureAsset(image, a.format)
	if err := a.resource.WriteContent(textureAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}
