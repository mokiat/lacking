package resource

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/graphics"
)

type CubeTexture struct {
	GFXTexture *graphics.CubeTexture
}

func NewCubeTextureOperator(locator Locator, gfxWorker *async.Worker) *CubeTextureOperator {
	return &CubeTextureOperator{
		locator:   locator,
		gfxWorker: gfxWorker,
	}
}

type CubeTextureOperator struct {
	locator   Locator
	gfxWorker *async.Worker
}

func (o *CubeTextureOperator) Allocator(uri string) Allocator {
	return AllocatorFunc(func(set *Set) (interface{}, error) {
		in, err := o.locator.Open(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to open cube texture asset %q: %w", uri, err)
		}
		defer in.Close()

		texAsset := new(asset.CubeTexture)
		if err := asset.DecodeCubeTexture(in, texAsset); err != nil {
			return nil, fmt.Errorf("failed to decode cube texture asset %q: %w", uri, err)
		}

		texture := &CubeTexture{
			GFXTexture: &graphics.CubeTexture{},
		}

		gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
			var dataFormat graphics.DataFormat
			switch texAsset.Format {
			case asset.DataFormatRGBA8:
				dataFormat = graphics.DataFormatRGBA8
			case asset.DataFormatRGBA32F:
				dataFormat = graphics.DataFormatRGBA32F
			default:
				return fmt.Errorf("unknown format: %d", dataFormat)
			}

			return texture.GFXTexture.Allocate(graphics.CubeTextureData{
				Dimension:      int32(texAsset.Dimension),
				Format:         dataFormat,
				FrontSideData:  texAsset.Sides[asset.TextureSideFront].Data,
				BackSideData:   texAsset.Sides[asset.TextureSideBack].Data,
				LeftSideData:   texAsset.Sides[asset.TextureSideLeft].Data,
				RightSideData:  texAsset.Sides[asset.TextureSideRight].Data,
				TopSideData:    texAsset.Sides[asset.TextureSideTop].Data,
				BottomSideData: texAsset.Sides[asset.TextureSideBottom].Data,
			})
		}))
		if err := gfxTask.Wait().Err; err != nil {
			return nil, fmt.Errorf("failed to allocate gfx cube texture: %w", err)
		}
		return texture, nil
	})
}

func (o *CubeTextureOperator) Releaser() Releaser {
	return ReleaserFunc(func(resource interface{}) error {
		texture := resource.(*CubeTexture)

		gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
			return texture.GFXTexture.Release()
		}))
		if err := gfxTask.Wait().Err; err != nil {
			return fmt.Errorf("failed to release gfx cube texture: %w", err)
		}
		return nil
	})
}
