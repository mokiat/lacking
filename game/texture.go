package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
)

// NOTE: The reason why we have a TwoDTexture wrapper in this package is
// to prevent the use from deleting the resource, as this is managed by
// the ResourceSet instead.
// Furthermore, in the future, we could have a `keep` option, which preserves
// the data and allows modification or using it as a hit-mask or similar.

type TwoDTexture struct {
	gfxTexture *graphics.TwoDTexture
}

type CubeTexture struct {
	gfxTexture *graphics.CubeTexture
}

func (r *ResourceSet) allocateTwoDTexture(resource asset.Resource) (*TwoDTexture, error) {
	texAsset := new(asset.TwoDTexture)
	if err := resource.ReadContent(texAsset); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	result := &TwoDTexture{}
	r.gfxWorker.Schedule(func() {
		gfxEngine := r.engine.Graphics()
		gfxTexture := gfxEngine.CreateTwoDTexture(graphics.TwoDTextureDefinition{
			Width:           int(texAsset.Width),
			Height:          int(texAsset.Height),
			Wrapping:        resolveWrapMode(texAsset.Wrapping),
			Filtering:       resolveFilter(texAsset.Filtering),
			GenerateMipmaps: texAsset.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !texAsset.Flags.Has(asset.TextureFlagLinear),
			DataFormat:      resolveDataFormat(texAsset.Format),
			InternalFormat:  resolveInternalFormat(texAsset.Format),
			Data:            texAsset.Data,
		})
		result.gfxTexture = gfxTexture
	})
	return result, nil
}

func (r *ResourceSet) releaseTwoDTexture(texture *TwoDTexture) {
	texture.gfxTexture.Delete()
}

func (r *ResourceSet) allocateCubeTexture(resource asset.Resource) (*CubeTexture, error) {
	texAsset := new(asset.CubeTexture)
	if err := resource.ReadContent(texAsset); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	result := &CubeTexture{}
	r.gfxWorker.Schedule(func() {
		gfxEngine := r.engine.Graphics()
		gfxTexture := gfxEngine.CreateCubeTexture(graphics.CubeTextureDefinition{
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
		result.gfxTexture = gfxTexture
	})
	return result, nil
}

func (r *ResourceSet) releaseCubeTexture(texture *CubeTexture) {
	texture.gfxTexture.Delete()
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
