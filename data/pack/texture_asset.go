package pack

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	newasset "github.com/mokiat/lacking/game/newasset"
)

func BuildTwoDTextureAsset(image *Image) *asset.TwoDTexture {
	return &asset.TwoDTexture{
		Width:  uint32(image.Width),
		Height: uint32(image.Height),
		Flags:  newasset.TextureFlag2D | newasset.TextureFlagMipmapping,
		Format: newasset.TexelFormatRGBA8,
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

func BuildCubeTextureAsset(image *CubeImage, format newasset.TexelFormat) *asset.CubeTexture {
	textureData := func(side CubeSide) []byte {
		switch format {
		case newasset.TexelFormatRGBA8:
			return image.RGBA8Data(side)
		case newasset.TexelFormatRGBA16F:
			return image.RGBA16FData(side)
		case newasset.TexelFormatRGBA32F:
			return image.RGBA32FData(side)
		default:
			panic(fmt.Errorf("unsupported format: %d", format))
		}
	}
	return &asset.CubeTexture{
		Dimension: uint32(image.Dimension),
		Flags:     newasset.TextureFlagCubeMap,
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
	format        newasset.TexelFormat
}

type SaveCubeTextureOption func(a *SaveCubeTextureAction)

func WithFormat(format newasset.TexelFormat) SaveCubeTextureOption {
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
