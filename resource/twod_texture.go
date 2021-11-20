package resource

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	gameasset "github.com/mokiat/lacking/game/asset"
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

func NewTwoDTextureOperator(delegate gameasset.Registry, gfxEngine graphics.Engine, gfxWorker *async.Worker) *TwoDTextureOperator {
	return &TwoDTextureOperator{
		delegate:  delegate,
		gfxEngine: gfxEngine,
		gfxWorker: gfxWorker,
	}
}

type TwoDTextureOperator struct {
	delegate  gameasset.Registry
	gfxEngine graphics.Engine
	gfxWorker *async.Worker
}

func (o *TwoDTextureOperator) Allocate(registry *Registry, id string) (interface{}, error) {
	texAsset := new(asset.TwoDTexture)
	if err := o.delegate.ReadContent(id, texAsset); err != nil {
		return nil, fmt.Errorf("failed to open twod texture asset %q: %w", id, err)
	}

	texture := &TwoDTexture{
		Name: id,
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
