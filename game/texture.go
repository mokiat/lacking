package game

import (
	"fmt"

	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
)

func (r *ResourceSet) allocateTexture(textureAsset asset.Texture) render.Texture {
	switch {
	case textureAsset.Flags.Has(asset.TextureFlag2D):
		return r.allocateTexture2D(textureAsset)
	case textureAsset.Flags.Has(asset.TextureFlagCubeMap):
		return r.allocateTextureCube(textureAsset)
	default:
		panic(fmt.Errorf("unsupported texture type (flags: %v)", textureAsset.Flags))
	}
}

func (r *ResourceSet) allocateTexture2D(textureAsset asset.Texture) render.Texture {
	var texture render.Texture
	r.gfxWorker.ScheduleVoid(func() {
		texture = r.renderAPI.CreateColorTexture2D(render.ColorTexture2DInfo{
			Width:           textureAsset.Width,
			Height:          textureAsset.Height,
			GenerateMipmaps: textureAsset.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !textureAsset.Flags.Has(asset.TextureFlagLinearSpace),
			Format:          resolveDataFormat(textureAsset.Format),
			Data:            textureAsset.Layers[0].Data,
		})
	}).Wait()
	return texture
}

func (r *ResourceSet) allocateTextureCube(textureAsset asset.Texture) render.Texture {
	var texture render.Texture
	r.gfxWorker.ScheduleVoid(func() {
		texture = r.renderAPI.CreateColorTextureCube(render.ColorTextureCubeInfo{
			Dimension:       textureAsset.Width,
			GenerateMipmaps: textureAsset.Flags.Has(asset.TextureFlagMipmapping),
			GammaCorrection: !textureAsset.Flags.Has(asset.TextureFlagLinearSpace),
			Format:          resolveDataFormat(textureAsset.Format),
			FrontSideData:   textureAsset.Layers[0].Data,
			BackSideData:    textureAsset.Layers[1].Data,
			LeftSideData:    textureAsset.Layers[2].Data,
			RightSideData:   textureAsset.Layers[3].Data,
			TopSideData:     textureAsset.Layers[4].Data,
			BottomSideData:  textureAsset.Layers[5].Data,
		})
	}).Wait()
	return texture
}
