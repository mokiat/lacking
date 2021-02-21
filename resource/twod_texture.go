package resource

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/graphics"
)

type TwoDTexture struct {
	GFXTexture *graphics.TwoDTexture
}

func NewTwoDTextureOperator(locator Locator, gfxWorker *async.Worker) *TwoDTextureOperator {
	return &TwoDTextureOperator{
		locator:   locator,
		gfxWorker: gfxWorker,
	}
}

type TwoDTextureOperator struct {
	locator   Locator
	gfxWorker *async.Worker
}

func (o *TwoDTextureOperator) Allocator(uri string) Allocator {
	return AllocatorFunc(func(set *Set) (interface{}, error) {
		in, err := o.locator.Open(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to open twod texture asset %q: %w", uri, err)
		}
		defer in.Close()

		texAsset := new(asset.TwoDTexture)
		if err := asset.DecodeTwoDTexture(in, texAsset); err != nil {
			return nil, fmt.Errorf("failed to decode twod texture asset %q: %w", uri, err)
		}

		texture := &TwoDTexture{
			GFXTexture: &graphics.TwoDTexture{},
		}

		gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
			return texture.GFXTexture.Allocate(graphics.TwoDTextureData{
				Width:  int32(texAsset.Width),
				Height: int32(texAsset.Height),
				Data:   texAsset.Data,
			})
		}))
		if err := gfxTask.Wait().Err; err != nil {
			return nil, fmt.Errorf("failed to allocate two dimensional gfx texture: %w", err)
		}
		return texture, nil
	})
}

func (o *TwoDTextureOperator) Releaser() Releaser {
	return ReleaserFunc(func(resource interface{}) error {
		texture := resource.(*TwoDTexture)

		gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
			return texture.GFXTexture.Release()
		}))
		if err := gfxTask.Wait().Err; err != nil {
			return fmt.Errorf("failed to release two dimensional gfx texture: %w", err)
		}
		return nil
	})
}
