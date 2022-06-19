package resource

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
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
	GFXTexture *graphics.TwoDTexture
}

func NewTwoDTextureOperator(delegate asset.Registry, gfxEngine *graphics.Engine) *TwoDTextureOperator {
	return &TwoDTextureOperator{
		delegate:  delegate,
		gfxEngine: gfxEngine,
	}
}

type TwoDTextureOperator struct {
	delegate  asset.Registry
	gfxEngine *graphics.Engine
}

func (o *TwoDTextureOperator) Allocate(registry *Registry, id string) (interface{}, error) {
	texAsset := new(asset.TwoDTexture)
	resource := o.delegate.ResourceByID(id)
	if resource == nil {
		return nil, fmt.Errorf("cannot find asset %q", id)
	}
	if err := resource.ReadContent(texAsset); err != nil {
		return nil, fmt.Errorf("failed to open twod texture asset %q: %w", id, err)
	}

	texture := &TwoDTexture{
		Name: id,
	}

	registry.ScheduleVoid(func() {
		definition := graphics.TwoDTextureDefinition{
			Width:           int(texAsset.Width),
			Height:          int(texAsset.Height),
			Wrapping:        resolveWrapMode(texAsset.Wrapping),
			Filtering:       resolveFilter(texAsset.Filtering),
			GenerateMipmaps: texAsset.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !texAsset.Flags.Has(asset.TextureFlagLinear),
			DataFormat:      graphics.DataFormatRGBA8,
			InternalFormat:  graphics.InternalFormatRGBA8,
			Data:            texAsset.Data,
		}
		texture.GFXTexture = o.gfxEngine.CreateTwoDTexture(definition)
	}).Wait()

	return texture, nil
}

func (o *TwoDTextureOperator) Release(registry *Registry, resource interface{}) error {
	texture := resource.(*TwoDTexture)

	registry.ScheduleVoid(func() {
		texture.GFXTexture.Delete()
	}).Wait()

	return nil
}

func resolveWrapMode(wrap asset.WrapMode) graphics.Wrap {
	switch wrap {
	case asset.WrapModeRepeat:
		return graphics.WrapRepeat
	case asset.WrapModeMirroredRepeat:
		return graphics.WrapMirroredRepat
	case asset.WrapModeClampToEdge:
		return graphics.WrapClampToEdge
	default:
		panic(fmt.Errorf("unknown wrap mode: %v", wrap))
	}
}
