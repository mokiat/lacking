package pack

import (
	"fmt"

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
	textureAsset := BuildTwoDTextureAsset(image)
	resource := a.registry.ResourceByID(a.id)
	if resource == nil {
		resource = a.registry.CreateResource("twod_texture", a.id)
	}
	if err := resource.WriteContent(textureAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	if err := a.registry.Save(); err != nil {
		return fmt.Errorf("error saving resources: %w", err)
	}
	return nil
}

type SaveCubeTextureAction struct {
	registry      gameasset.Registry
	id            string
	imageProvider CubeImageProvider
	format        gameasset.TexelFormat
}

type SaveCubeTextureOption func(a *SaveCubeTextureAction)

func WithFormat(format gameasset.TexelFormat) SaveCubeTextureOption {
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
		case gameasset.TexelFormatRGBA8:
			return texture.RGBA8Data(side)
		case gameasset.TexelFormatRGBA16F:
			return texture.RGBA16FData(side)
		case gameasset.TexelFormatRGBA32F:
			return texture.RGBA32FData(side)
		default:
			panic(fmt.Errorf("unsupported format: %d", a.format))
		}
	}

	textureAsset := &gameasset.CubeTexture{
		Dimension: uint16(texture.Dimension),
		Filtering: gameasset.FilterModeLinear,
		Flags:     gameasset.TextureFlagNone,
		Format:    gameasset.TexelFormat(a.format),
	}
	textureAsset.FrontSide = gameasset.CubeTextureSide{
		Data: textureData(CubeSideFront),
	}
	textureAsset.BackSide = gameasset.CubeTextureSide{
		Data: textureData(CubeSideRear),
	}
	textureAsset.LeftSide = gameasset.CubeTextureSide{
		Data: textureData(CubeSideLeft),
	}
	textureAsset.RightSide = gameasset.CubeTextureSide{
		Data: textureData(CubeSideRight),
	}
	textureAsset.TopSide = gameasset.CubeTextureSide{
		Data: textureData(CubeSideTop),
	}
	textureAsset.BottomSide = gameasset.CubeTextureSide{
		Data: textureData(CubeSideBottom),
	}

	resource := a.registry.ResourceByID(a.id)
	if resource == nil {
		resource = a.registry.CreateResource("cube_texture", a.id)
	}
	if err := resource.WriteContent(textureAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	if err := a.registry.Save(); err != nil {
		return fmt.Errorf("error saving resources: %w", err)
	}
	return nil
}

func BuildTwoDTextureAsset(image *Image) *gameasset.TwoDTexture {
	textureAsset := &gameasset.TwoDTexture{
		Width:     uint16(image.Width),
		Height:    uint16(image.Height),
		Wrapping:  gameasset.WrapModeRepeat,
		Filtering: gameasset.FilterModeLinear,
		Flags:     gameasset.TextureFlagMipmapping,
		Format:    gameasset.TexelFormatRGBA8,
		Data:      image.RGBA8Data(),
	}
	return textureAsset
}
