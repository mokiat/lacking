package resource

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/graphics"
)

const TwoDTextureTypeName = TypeName("twod_texture")

func InjectTwoDTexture(target **TwoDTexture) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*TwoDTexture)
	}
}

type TwoDTexture struct {
	Name       string
	GFXTexture *graphics.TwoDTexture
}

func NewTwoDTextureOperator(locator Locator, gfxWorker *graphics.Worker) *TwoDTextureOperator {
	return &TwoDTextureOperator{
		locator:   locator,
		gfxWorker: gfxWorker,
	}
}

type TwoDTextureOperator struct {
	locator   Locator
	gfxWorker *graphics.Worker
}

func (o *TwoDTextureOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "textures", "twod", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open twod texture asset %q: %w", name, err)
	}
	defer in.Close()

	texAsset := new(asset.TwoDTexture)
	if err := asset.DecodeTwoDTexture(in, texAsset); err != nil {
		return nil, fmt.Errorf("failed to decode twod texture asset %q: %w", name, err)
	}

	texture := &TwoDTexture{
		Name:       name,
		GFXTexture: &graphics.TwoDTexture{},
	}

	gfxTask := o.gfxWorker.Schedule(func() error {
		return texture.GFXTexture.Allocate(graphics.TwoDTextureData{
			Width:  int32(texAsset.Width),
			Height: int32(texAsset.Height),
			Data:   texAsset.Data,
		})
	})
	if err := gfxTask.Wait(); err != nil {
		return nil, fmt.Errorf("failed to allocate two dimensional gfx texture: %w", err)
	}
	return texture, nil
}

func (o *TwoDTextureOperator) Release(registry *Registry, resource interface{}) error {
	texture := resource.(*TwoDTexture)

	gfxTask := o.gfxWorker.Schedule(func() error {
		return texture.GFXTexture.Release()
	})
	if err := gfxTask.Wait(); err != nil {
		return fmt.Errorf("failed to release two dimensional gfx texture: %w", err)
	}
	return nil
}
