package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
)

type TwoDTextureAssetBuilder struct {
	Asset
	imageProvider ImageProvider
}

func (b *TwoDTextureAssetBuilder) WithImage(imageProvider ImageProvider) *TwoDTextureAssetBuilder {
	b.imageProvider = imageProvider
	return b
}

func (b *TwoDTextureAssetBuilder) Build() error {
	img, err := b.imageProvider.Image()
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	texture := &asset.TwoDTexture{
		Width:  uint16(img.Width()),
		Height: uint16(img.Height()),
		Data:   img.RGBAData(),
	}

	file, err := b.CreateFile()
	if err != nil {
		return err
	}
	defer file.Close()

	if err := asset.EncodeTwoDTexture(file, texture); err != nil {
		return fmt.Errorf("failed to encode twod texture: %w", err)
	}
	return nil
}
