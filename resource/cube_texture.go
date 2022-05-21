package resource

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
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
	GFXTexture *graphics.CubeTexture
}

func NewCubeTextureOperator(delegate asset.Registry, gfxEngine *graphics.Engine) *CubeTextureOperator {
	return &CubeTextureOperator{
		delegate:  delegate,
		gfxEngine: gfxEngine,
	}
}

type CubeTextureOperator struct {
	delegate  asset.Registry
	gfxEngine *graphics.Engine
}

func (o *CubeTextureOperator) Allocate(registry *Registry, id string) (interface{}, error) {
	texAsset := new(asset.CubeTexture)
	if err := o.delegate.ReadContent(id, texAsset); err != nil {
		return nil, fmt.Errorf("failed to open cube texture asset %q: %w", id, err)
	}

	texture := &CubeTexture{
		Name: id,
	}

	registry.ScheduleVoid(func() {
		texture.GFXTexture = o.gfxEngine.CreateCubeTexture(graphics.CubeTextureDefinition{
			Dimension:      int(texAsset.Dimension),
			Filtering:      resolveFilter(texAsset.Filtering),
			DataFormat:     resolveDataFormat(texAsset.Format),
			InternalFormat: resolveInternalFormat(texAsset.Format),
			FrontSideData:  texAsset.FrontSide.Data,
			BackSideData:   texAsset.BackSide.Data,
			LeftSideData:   texAsset.LeftSide.Data,
			RightSideData:  texAsset.RightSide.Data,
			TopSideData:    texAsset.TopSide.Data,
			BottomSideData: texAsset.BottomSide.Data,
		})
	}).Wait()

	return texture, nil
}

func (o *CubeTextureOperator) Release(registry *Registry, resource interface{}) error {
	texture := resource.(*CubeTexture)

	registry.ScheduleVoid(func() {
		texture.GFXTexture.Delete()
	}).Wait()

	return nil
}

func resolveFilter(filter asset.FilterMode) graphics.Filter {
	switch filter {
	case asset.FilterModeNearest:
		return graphics.FilterNearest
	case asset.FilterModeLinear:
		return graphics.FilterLinear
	case asset.FilterModeAnisotropic:
		return graphics.FilterAnisotropic
	default:
		panic(fmt.Errorf("unknown filter mode: %v", filter))
	}
}

func resolveDataFormat(format asset.TexelFormat) graphics.DataFormat {
	// FIXME: Support other formats as well
	switch format {
	case asset.TexelFormatRGBA8:
		return graphics.DataFormatRGBA8
	case asset.TexelFormatRGBA16F:
		return graphics.DataFormatRGBA16F
	case asset.TexelFormatRGBA32F:
		return graphics.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}

func resolveInternalFormat(format asset.TexelFormat) graphics.InternalFormat {
	// FIXME: Support other formats as well
	switch format {
	case asset.TexelFormatRGBA8:
		return graphics.InternalFormatRGBA8
	case asset.TexelFormatRGBA16F:
		return graphics.InternalFormatRGBA16F
	case asset.TexelFormatRGBA32F:
		return graphics.InternalFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}
