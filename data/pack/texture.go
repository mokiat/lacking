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

type CubeTextureAssetBuilder struct {
	Asset
	frontImageProvider  ImageProvider
	backImageProvider   ImageProvider
	leftImageProvider   ImageProvider
	rightImageProvider  ImageProvider
	topImageProvider    ImageProvider
	bottomImageProvider ImageProvider
	dimension           int
}

func (b *CubeTextureAssetBuilder) WithFrontImage(imageProvider ImageProvider) *CubeTextureAssetBuilder {
	b.frontImageProvider = imageProvider
	return b
}

func (b *CubeTextureAssetBuilder) WithBackImage(imageProvider ImageProvider) *CubeTextureAssetBuilder {
	b.backImageProvider = imageProvider
	return b
}

func (b *CubeTextureAssetBuilder) WithLeftImage(imageProvider ImageProvider) *CubeTextureAssetBuilder {
	b.leftImageProvider = imageProvider
	return b
}

func (b *CubeTextureAssetBuilder) WithRightImage(imageProvider ImageProvider) *CubeTextureAssetBuilder {
	b.rightImageProvider = imageProvider
	return b
}

func (b *CubeTextureAssetBuilder) WithTopImage(imageProvider ImageProvider) *CubeTextureAssetBuilder {
	b.topImageProvider = imageProvider
	return b
}

func (b *CubeTextureAssetBuilder) WithBottomImage(imageProvider ImageProvider) *CubeTextureAssetBuilder {
	b.bottomImageProvider = imageProvider
	return b
}

func (b *CubeTextureAssetBuilder) WithDimension(dimension int) *CubeTextureAssetBuilder {
	b.dimension = dimension
	return b
}

func (b *CubeTextureAssetBuilder) Build() error {
	frontImg, err := b.frontImageProvider.Image()
	if err != nil {
		return fmt.Errorf("failed to get front image: %w", err)
	}
	if !frontImg.IsSquare() {
		return fmt.Errorf("front image is not a square")
	}

	backImg, err := b.backImageProvider.Image()
	if err != nil {
		return fmt.Errorf("failed to get back image: %w", err)
	}
	if !backImg.IsSquare() {
		return fmt.Errorf("back image is not a square")
	}

	leftImg, err := b.leftImageProvider.Image()
	if err != nil {
		return fmt.Errorf("failed to get left image: %w", err)
	}
	if !leftImg.IsSquare() {
		return fmt.Errorf("left image is not a square")
	}

	rightImg, err := b.rightImageProvider.Image()
	if err != nil {
		return fmt.Errorf("failed to get right image: %w", err)
	}
	if !rightImg.IsSquare() {
		return fmt.Errorf("right image is not a square")
	}

	topImg, err := b.topImageProvider.Image()
	if err != nil {
		return fmt.Errorf("failed to get top image: %w", err)
	}
	if !topImg.IsSquare() {
		return fmt.Errorf("top image is not a square")
	}

	bottomImg, err := b.bottomImageProvider.Image()
	if err != nil {
		return fmt.Errorf("failed to get bottom image: %w", err)
	}
	if !bottomImg.IsSquare() {
		return fmt.Errorf("bottom image is not a square")
	}

	areSameDimension := frontImg.Width() == backImg.Width() &&
		frontImg.Width() == leftImg.Width() &&
		frontImg.Width() == rightImg.Width() &&
		frontImg.Width() == topImg.Width() &&
		frontImg.Width() == bottomImg.Width()
	if !areSameDimension {
		return fmt.Errorf("images are not of the same size")
	}

	if b.dimension > 0 {
		frontImg.Scale(b.dimension, b.dimension)
		backImg.Scale(b.dimension, b.dimension)
		leftImg.Scale(b.dimension, b.dimension)
		rightImg.Scale(b.dimension, b.dimension)
		topImg.Scale(b.dimension, b.dimension)
		bottomImg.Scale(b.dimension, b.dimension)
	}

	texture := &asset.CubeTexture{
		Dimension: uint16(frontImg.Width()),
	}
	texture.Sides[asset.TextureSideFront] = asset.CubeTextureSide{
		Data: frontImg.RGBAData(),
	}
	texture.Sides[asset.TextureSideBack] = asset.CubeTextureSide{
		Data: backImg.RGBAData(),
	}
	texture.Sides[asset.TextureSideLeft] = asset.CubeTextureSide{
		Data: leftImg.RGBAData(),
	}
	texture.Sides[asset.TextureSideRight] = asset.CubeTextureSide{
		Data: rightImg.RGBAData(),
	}
	texture.Sides[asset.TextureSideTop] = asset.CubeTextureSide{
		Data: topImg.RGBAData(),
	}
	texture.Sides[asset.TextureSideBottom] = asset.CubeTextureSide{
		Data: bottomImg.RGBAData(),
	}

	file, err := b.CreateFile()
	if err != nil {
		return err
	}
	defer file.Close()

	if err := asset.EncodeCubeTexture(file, texture); err != nil {
		return fmt.Errorf("failed to encode cube texture: %w", err)
	}
	return nil
}
