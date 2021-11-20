package resource

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/game/graphics"
)

const CubeTextureTypeName = TypeName("cube_texture")

func InjectCubeTexture(target **CubeTexture) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*CubeTexture)
	}
}

type CubeTexture struct {
	Name       string
	GFXTexture graphics.CubeTexture
}

func NewCubeTextureOperator(locator Locator, gfxEngine graphics.Engine, gfxWorker *async.Worker) *CubeTextureOperator {
	return &CubeTextureOperator{
		locator:   locator,
		gfxEngine: gfxEngine,
		gfxWorker: gfxWorker,
	}
}

type CubeTextureOperator struct {
	locator   Locator
	gfxEngine graphics.Engine
	gfxWorker *async.Worker
}

func (o *CubeTextureOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "textures", "cube", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open cube texture asset %q: %w", name, err)
	}
	defer in.Close()

	texAsset := new(asset.CubeTexture)
	if err := asset.Decode(in, texAsset); err != nil {
		return nil, fmt.Errorf("failed to decode cube texture asset %q: %w", name, err)
	}

	texture := &CubeTexture{
		Name: name,
	}

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		definition := graphics.CubeTextureDefinition{
			Dimension:      int(texAsset.Dimension),
			WrapS:          graphics.WrapClampToEdge,
			WrapT:          graphics.WrapClampToEdge,
			MinFilter:      graphics.FilterLinear,
			MagFilter:      graphics.FilterLinear,
			FrontSideData:  texAsset.Sides[asset.TextureSideFront].Data,
			BackSideData:   texAsset.Sides[asset.TextureSideBack].Data,
			LeftSideData:   texAsset.Sides[asset.TextureSideLeft].Data,
			RightSideData:  texAsset.Sides[asset.TextureSideRight].Data,
			TopSideData:    texAsset.Sides[asset.TextureSideTop].Data,
			BottomSideData: texAsset.Sides[asset.TextureSideBottom].Data,
		}
		switch texAsset.Format {
		case asset.TexelFormatRGBA8:
			definition.DataFormat = graphics.DataFormatRGBA8
			definition.InternalFormat = graphics.InternalFormatRGBA8
		case asset.TexelFormatRGBA32F:
			definition.DataFormat = graphics.DataFormatRGBA32F
			definition.InternalFormat = graphics.InternalFormatRGBA32F
		default:
			return fmt.Errorf("unknown format: %d", texAsset.Format)
		}
		texture.GFXTexture = o.gfxEngine.CreateCubeTexture(definition)
		return nil
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return nil, fmt.Errorf("failed to allocate gfx cube texture: %w", err)
	}
	return texture, nil
}

func (o *CubeTextureOperator) Release(registry *Registry, resource interface{}) error {
	texture := resource.(*CubeTexture)

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		texture.GFXTexture.Delete()
		return nil
	}))

	if err := gfxTask.Wait().Err; err != nil {
		return fmt.Errorf("failed to release gfx cube texture: %w", err)
	}
	return nil
}
