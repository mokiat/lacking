package game

import (
	"fmt"

	"github.com/mokiat/gog"
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
	s.gfxWorker.Schedule(func() {
		texture := s.renderAPI.CreateColorTexture2D(render.ColorTexture2DInfo{
			GenerateMipmaps: assetTexture.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(asset.TextureFlagLinearSpace),
			Format:          s.resolveDataFormat(assetTexture.Format),
			MipmapLayers: gog.Map(assetTexture.MipmapLayers, func(layer asset.MipmapLayer) render.Mipmap2DLayer {
				return render.Mipmap2DLayer{
					Width:  layer.Width,
					Height: layer.Height,
					Data:   layer.Layers[0].Data,
				}
			}),
		})
		promise.Deliver(texture)
	})
	return promise
}

func (s *ResourceSet) allocateTextureCube(assetTexture asset.Texture) async.Promise[render.Texture] {
	promise := async.NewPromise[render.Texture]()
	s.gfxWorker.Schedule(func() {
		texture := s.renderAPI.CreateColorTextureCube(render.ColorTextureCubeInfo{
			GenerateMipmaps: assetTexture.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(asset.TextureFlagLinearSpace),
			Format:          s.resolveDataFormat(assetTexture.Format),
			MipmapLayers: gog.Map(assetTexture.MipmapLayers, func(layer asset.MipmapLayer) render.MipmapCubeLayer {
				return render.MipmapCubeLayer{
					Dimension:      layer.Width,
					FrontSideData:  layer.Layers[0].Data,
					BackSideData:   layer.Layers[1].Data,
					LeftSideData:   layer.Layers[2].Data,
					RightSideData:  layer.Layers[3].Data,
					TopSideData:    layer.Layers[4].Data,
					BottomSideData: layer.Layers[5].Data,
				}
			}),
		})
		promise.Deliver(texture)
	})
	return promise
}
