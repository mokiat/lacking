package resource

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/graphics"
)

const CubeTextureTypeName = TypeName("cube_texture")

func InjectCubeTexture(target **CubeTexture) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*CubeTexture)
	}
}

type CubeTexture struct {
	Name       string
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

func (o *CubeTextureOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "textures", "cube", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open cube texture asset %q: %w", name, err)
	}
	defer in.Close()

	texAsset := new(asset.CubeTexture)
	if err := asset.DecodeCubeTexture(in, texAsset); err != nil {
		return nil, fmt.Errorf("failed to decode cube texture asset %q: %w", name, err)
	}

	texture := &CubeTexture{
		Name:       name,
		GFXTexture: &graphics.CubeTexture{},
	}

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		return texture.GFXTexture.Allocate(graphics.CubeTextureData{
			Dimension:      int32(texAsset.Dimension),
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
}

func (o *CubeTextureOperator) Release(registry *Registry, resource interface{}) error {
	texture := resource.(*CubeTexture)

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		return texture.GFXTexture.Release()
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return fmt.Errorf("failed to release gfx cube texture: %w", err)
	}
	return nil
}
