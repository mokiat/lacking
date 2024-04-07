package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	newasset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
)

func (r *ResourceSet) allocateTwoDTexture(texAsset *asset.TwoDTexture) render.Texture {
	renderAPI := r.engine.Graphics().API()

	var texture render.Texture
	r.gfxWorker.ScheduleVoid(func() {
		texture = renderAPI.CreateColorTexture2D(render.ColorTexture2DInfo{
			Width:           uint32(texAsset.Width),
			Height:          uint32(texAsset.Height),
			GenerateMipmaps: texAsset.Flags.Has(newasset.TextureFlagMipmapping),
			GammaCorrection: !texAsset.Flags.Has(newasset.TextureFlagLinearSpace),
			Format:          resolveDataFormat(texAsset.Format),
			Data:            texAsset.Data,
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
