package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	newasset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
)

func (r *ResourceSet) allocateTexture(textureAsset newasset.Texture) render.Texture {
	switch {
	case textureAsset.Flags.Has(newasset.TextureFlag2D):
		return r.allocateTexture2D(textureAsset)
	case textureAsset.Flags.Has(newasset.TextureFlagCubeMap):
		return r.allocateTextureCube(textureAsset)
	default:
		panic(fmt.Errorf("unsupported texture type (flags: %v)", textureAsset.Flags))
	}
}

func (r *ResourceSet) allocateTexture2D(textureAsset newasset.Texture) render.Texture {
	var texture render.Texture
	r.gfxWorker.ScheduleVoid(func() {
		texture = r.renderAPI.CreateColorTexture2D(render.ColorTexture2DInfo{
			Width:           textureAsset.Width,
			Height:          textureAsset.Height,
			GenerateMipmaps: textureAsset.Flags.Has(newasset.TextureFlagMipmapping),
			GammaCorrection: !textureAsset.Flags.Has(newasset.TextureFlagLinearSpace),
			Format:          resolveDataFormat(textureAsset.Format),
			Data:            textureAsset.Layers[0].Data,
		})
	}).Wait()
	return texture
}

func (r *ResourceSet) allocateTextureCube(textureAsset newasset.Texture) render.Texture {
	var texture render.Texture
	r.gfxWorker.ScheduleVoid(func() {
		texture = r.renderAPI.CreateColorTextureCube(render.ColorTextureCubeInfo{
			Dimension:       textureAsset.Width,
			GenerateMipmaps: textureAsset.Flags.Has(newasset.TextureFlagMipmapping),
			GammaCorrection: !textureAsset.Flags.Has(newasset.TextureFlagLinearSpace),
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

func (r *ResourceSet) allocateCubeTexture(resource asset.Resource) (render.Texture, error) {
	renderAPI := r.engine.Graphics().API()

	texAsset := new(asset.CubeTexture)
	ioTask := func() error {
		return resource.ReadContent(texAsset)
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}

	var texture render.Texture
	r.gfxWorker.ScheduleVoid(func() {
		texture = renderAPI.CreateColorTextureCube(render.ColorTextureCubeInfo{
			Dimension:       uint32(texAsset.Dimension),
			GenerateMipmaps: texAsset.Flags.Has(newasset.TextureFlagMipmapping),
			GammaCorrection: !texAsset.Flags.Has(newasset.TextureFlagLinearSpace),
			Format:          resolveDataFormat(texAsset.Format),
			FrontSideData:   texAsset.FrontSide.Data,
			BackSideData:    texAsset.BackSide.Data,
			LeftSideData:    texAsset.LeftSide.Data,
			RightSideData:   texAsset.RightSide.Data,
			TopSideData:     texAsset.TopSide.Data,
			BottomSideData:  texAsset.BottomSide.Data,
		})
	}).Wait()
	return texture, nil
}
