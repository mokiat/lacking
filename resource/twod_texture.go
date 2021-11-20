package resource

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/game/graphics"
)

const TwoDTextureTypeName = TypeName("twod_texture")

func InjectTwoDTexture(target **TwoDTexture) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*TwoDTexture)
	}
}

type TwoDTexture struct {
	Name       string
	GFXTexture graphics.TwoDTexture
}

func NewTwoDTextureOperator(locator Locator, gfxEngine graphics.Engine, gfxWorker *async.Worker) *TwoDTextureOperator {
	return &TwoDTextureOperator{
		locator:   locator,
		gfxEngine: gfxEngine,
		gfxWorker: gfxWorker,
	}
}

type TwoDTextureOperator struct {
	locator   Locator
	gfxEngine graphics.Engine
	gfxWorker *async.Worker
}

func (o *TwoDTextureOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "textures", "twod", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open twod texture asset %q: %w", name, err)
	}
	defer in.Close()

	texAsset := new(asset.TwoDTexture)
	if err := asset.Decode(in, texAsset); err != nil {
		return nil, fmt.Errorf("failed to decode twod texture asset %q: %w", name, err)
	}

	texture := &TwoDTexture{
		Name: name,
	}

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		definition := graphics.TwoDTextureDefinition{
			Width:           int(texAsset.Width),
			Height:          int(texAsset.Height),
			WrapS:           graphics.WrapRepeat,
			WrapT:           graphics.WrapRepeat,
			MinFilter:       graphics.FilterLinearMipmapLinear,
			MagFilter:       graphics.FilterLinear,
			UseAnisotropy:   true,
			GenerateMipmaps: true,
			DataFormat:      graphics.DataFormatRGBA8,
			InternalFormat:  graphics.InternalFormatRGBA8,
			Data:            texAsset.Data,
		}
		texture.GFXTexture = o.gfxEngine.CreateTwoDTexture(definition)
		return nil
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return nil, fmt.Errorf("failed to allocate two dimensional gfx texture: %w", err)
	}
	return texture, nil
}

func (o *TwoDTextureOperator) Release(registry *Registry, resource interface{}) error {
	texture := resource.(*TwoDTexture)

	gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
		texture.GFXTexture.Delete()
		return nil
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return fmt.Errorf("failed to release two dimensional gfx texture: %w", err)
	}
	return nil
}

// func convertWrapMode(wrap asset.WrapMode) graphics.Wrap {
// 	switch wrap {
// 	case asset.WrapModeDefault:
// 		return graphics.WrapClampToEdge
// 	case asset.WrapModeRepeat:
// 		return graphics.WrapRepeat
// 	case asset.WrapModeMirroredRepeat:
// 		return graphics.WrapRepeat // FIXME
// 	case asset.WrapModeClampToEdge:
// 		return graphics.WrapClampToEdge
// 	case asset.WrapModeMirroredClampToEdge:
// 		return graphics.WrapClampToEdge // FIXME
// 	default:
// 		panic(fmt.Errorf("unknown wrap mode: %v", wrap))
// 	}
// }
