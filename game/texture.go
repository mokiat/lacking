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

func (r *ResourceSet) loadTwoDTexture(resource asset.Resource) (*TwoDTexture, error) {
	texAsset := new(asset.TwoDTexture)

	ioTask := func() error {
		return resource.ReadContent(texAsset)
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}

	return r.allocateTwoDTexture(texAsset), nil
}

func (r *ResourceSet) allocateTwoDTexture(texAsset *asset.TwoDTexture) *TwoDTexture {
	var gfxTexture *graphics.TwoDTexture
	r.gfxWorker.ScheduleVoid(func() {
		gfxEngine := r.engine.Graphics()
		gfxTexture = gfxEngine.CreateTwoDTexture(graphics.TwoDTextureDefinition{
			Width:           int(texAsset.Width),
			Height:          int(texAsset.Height),
			GenerateMipmaps: texAsset.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !texAsset.Flags.Has(asset.TextureFlagLinear),
			DataFormat:      resolveDataFormat(texAsset.Format),
			InternalFormat:  resolveInternalFormat(texAsset.Format),
			Data:            texAsset.Data,
		})
	}).Wait()
	return &TwoDTexture{
		gfxTexture: gfxTexture,
	}
}

func (r *ResourceSet) releaseTwoDTexture(texture *TwoDTexture) {
	texture.gfxTexture.Delete()
}

func (r *ResourceSet) allocateCubeTexture(resource asset.Resource) (*CubeTexture, error) {
	texAsset := new(asset.CubeTexture)

	ioTask := func() error {
		return resource.ReadContent(texAsset)
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}

	var gfxTexture *graphics.CubeTexture
	r.gfxWorker.ScheduleVoid(func() {
		gfxEngine := r.engine.Graphics()
		gfxTexture = gfxEngine.CreateCubeTexture(graphics.CubeTextureDefinition{
			Dimension:      int(texAsset.Dimension),
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
	return &CubeTexture{
		gfxTexture: gfxTexture,
	}, nil
}

func (r *ResourceSet) releaseCubeTexture(texture *CubeTexture) {
	texture.gfxTexture.Delete()
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
