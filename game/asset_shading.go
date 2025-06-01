package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/render"
	"golang.org/x/sync/errgroup"
)

func (l *AssetLoader) ResolveShader(assetShader dto.Shader) (Identifiable[*graphics.Shader], error) {
	var shader *graphics.Shader

	err := l.ScheduleMain(func(engine *Engine) error {
		gfxEngine := engine.Graphics()
		shader = gfxEngine.CreateShader(graphics.ShaderInfo{
			ShaderType: l.resolveShaderType(assetShader.ShaderType),
			SourceCode: assetShader.SourceCode,
		})
		return nil
	}).Wait()

	return Identifiable[*graphics.Shader]{
		ID:    assetShader.ID,
		Value: shader,
	}, err
}

func (l *AssetLoader) ResolveShaders(assetShaders []dto.Shader) (IdentifiableList[*graphics.Shader], error) {
	shaders := make(IdentifiableList[*graphics.Shader], len(assetShaders))
	var group errgroup.Group
	for i, assetShader := range assetShaders {
		group.Go(func() error {
			shader, err := l.ResolveShader(assetShader)
			shaders[i] = shader
			return err
		})
	}
	return shaders, group.Wait()
}

func (l *AssetLoader) resolveShaderType(assetType dto.ShaderType) graphics.ShaderType {
	switch assetType {
	case dto.ShaderTypeGeometry:
		return graphics.ShaderTypeGeometry
	case dto.ShaderTypeShadow:
		return graphics.ShaderTypeShadow
	case dto.ShaderTypeForward:
		return graphics.ShaderTypeForward
	case dto.ShaderTypeSky:
		return graphics.ShaderTypeSky
	case dto.ShaderTypePostprocess:
		return graphics.ShaderTypePostprocess
	default:
		panic(fmt.Errorf("unsupported shader type: %d", assetType))
	}
}

func (l *AssetLoader) ResolveTexture(assetTexture dto.Texture) (Identifiable[render.Texture], error) {
	switch {
	case assetTexture.Flags.Has(dto.TextureFlag2D):
		return l.ResolveTexture2D(assetTexture)
	case assetTexture.Flags.Has(dto.TextureFlagCubeMap):
		return l.ResolveTextureCube(assetTexture)
	default:
		return Identifiable[render.Texture]{}, fmt.Errorf("unsupported texture type (flags: %v)", assetTexture.Flags)
	}
}

func (l *AssetLoader) ResolveTexture2D(assetTexture dto.Texture) (Identifiable[render.Texture], error) {
	var texture render.Texture

	l.ScheduleMain(func(engine *Engine) error {
		renderAPI := engine.Graphics().API()
		texture = renderAPI.CreateColorTexture2D(render.ColorTexture2DInfo{
			GenerateMipmaps: assetTexture.Flags.Has(dto.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(dto.TextureFlagLinearSpace),
			Format:          l.resolveDataFormat(assetTexture.Format),
			MipmapLayers: gog.Map(assetTexture.MipmapLayers, func(layer dto.MipmapLayer) render.Mipmap2DLayer {
				return render.Mipmap2DLayer{
					Width:  layer.Width,
					Height: layer.Height,
					Data:   layer.Layers[0].Data,
				}
			}),
		})
		return nil
	}).Wait()

	return Identifiable[render.Texture]{
		ID:    assetTexture.ID,
		Value: texture,
	}, nil
}

func (l *AssetLoader) ResolveTextureCube(assetTexture dto.Texture) (Identifiable[render.Texture], error) {
	var texture render.Texture

	l.ScheduleMain(func(engine *Engine) error {
		renderAPI := engine.Graphics().API()
		texture = renderAPI.CreateColorTextureCube(render.ColorTextureCubeInfo{
			GenerateMipmaps: assetTexture.Flags.Has(dto.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(dto.TextureFlagLinearSpace),
			Format:          l.resolveDataFormat(assetTexture.Format),
			MipmapLayers: gog.Map(assetTexture.MipmapLayers, func(layer dto.MipmapLayer) render.MipmapCubeLayer {
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
		return nil
	}).Wait()

	return Identifiable[render.Texture]{
		ID:    assetTexture.ID,
		Value: texture,
	}, nil
}

func (l *AssetLoader) ResolveTextures(assetTextures []dto.Texture) (IdentifiableList[render.Texture], error) {
	textures := make(IdentifiableList[render.Texture], len(assetTextures))
	var group errgroup.Group
	for i, assetTexture := range assetTextures {
		group.Go(func() error {
			texture, err := l.ResolveTexture(assetTexture)
			textures[i] = texture
			return err
		})
	}
	return textures, group.Wait()
}

func (l *AssetLoader) resolveDataFormat(format dto.TexelFormat) render.DataFormat {
	switch format {
	case dto.TexelFormatRGBA8:
		return render.DataFormatRGBA8
	case dto.TexelFormatRGBA16F:
		return render.DataFormatRGBA16F
	case dto.TexelFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}
