package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertTexture(assetTexture asset.Texture) async.Promise[render.Texture] {
	switch {
	case assetTexture.Flags.Has(asset.TextureFlag2D):
		return s.allocateTexture2D(assetTexture)
	case assetTexture.Flags.Has(asset.TextureFlagCubeMap):
		return s.allocateTextureCube(assetTexture)
	default:
		err := fmt.Errorf("unsupported texture type (flags: %v)", assetTexture.Flags)
		return async.NewFailedPromise[render.Texture](err)
	}
}

func (s *ResourceSet) allocateTexture2D(assetTexture asset.Texture) async.Promise[render.Texture] {
	promise := async.NewPromise[render.Texture]()
	s.gfxWorker.ScheduleVoid(func() {
		texture := s.renderAPI.CreateColorTexture2D(render.ColorTexture2DInfo{
			Width:           assetTexture.Width,
			Height:          assetTexture.Height,
			GenerateMipmaps: assetTexture.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(asset.TextureFlagLinearSpace),
			Format:          s.resolveDataFormat(assetTexture.Format),
			Data:            assetTexture.Layers[0].Data,
		})
		promise.Deliver(texture)
	})
	return promise
}

func (s *ResourceSet) allocateTextureCube(assetTexture asset.Texture) async.Promise[render.Texture] {
	promise := async.NewPromise[render.Texture]()
	s.gfxWorker.ScheduleVoid(func() {
		texture := s.renderAPI.CreateColorTextureCube(render.ColorTextureCubeInfo{
			Dimension:       assetTexture.Width,
			GenerateMipmaps: assetTexture.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(asset.TextureFlagLinearSpace),
			Format:          s.resolveDataFormat(assetTexture.Format),
			FrontSideData:   assetTexture.Layers[0].Data,
			BackSideData:    assetTexture.Layers[1].Data,
			LeftSideData:    assetTexture.Layers[2].Data,
			RightSideData:   assetTexture.Layers[3].Data,
			TopSideData:     assetTexture.Layers[4].Data,
			BottomSideData:  assetTexture.Layers[5].Data,
		})
		promise.Deliver(texture)
	})
	return promise
}
