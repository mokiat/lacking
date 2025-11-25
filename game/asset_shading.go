package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/render"
	"golang.org/x/sync/errgroup"
)

// LoadShader resolves a shader from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadShader(loader *AssetLoader, assetShader dto.Shader) (Identifiable[*graphics.Shader], error) {
	var shader *graphics.Shader

	allocateShader := func() error {
		gfxEngine := loader.Engine().Graphics()
		shader = gfxEngine.CreateShader(graphics.ShaderInfo{
			ShaderType: resolveShaderType(assetShader.ShaderType),
			SourceCode: assetShader.SourceCode,
		})
		return nil
	}
	if err := loader.ScheduleMain(allocateShader).Wait(); err != nil {
		return Identifiable[*graphics.Shader]{}, err
	}

	return Identifiable[*graphics.Shader]{
		ID:    assetShader.ID,
		Value: shader,
	}, nil
}

// LoadShaders resolves a list of shaders from the given asset shaders.
//
// This is a blocking operation and should be called from a worker thread.
func LoadShaders(loader *AssetLoader, assetShaders []dto.Shader) (IdentifiableList[*graphics.Shader], error) {
	shaders := make(IdentifiableList[*graphics.Shader], len(assetShaders))
	var group errgroup.Group
	for i, assetShader := range assetShaders {
		group.Go(func() error {
			shader, err := LoadShader(loader, assetShader)
			shaders[i] = shader
			return err
		})
	}
	return shaders, group.Wait()
}

// UnloadShader unloads a shader from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadShader(loader *AssetLoader, idShader Identifiable[*graphics.Shader]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadShaders unloads a list of shaders from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadShaders(loader *AssetLoader, idShaders IdentifiableList[*graphics.Shader]) error {
	for _, idShader := range idShaders {
		if err := UnloadShader(loader, idShader); err != nil {
			return err
		}
	}
	return nil
}

// LoadTexture resolves a texture from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadTexture(loader *AssetLoader, assetTexture dto.Texture) (Identifiable[render.Texture], error) {
	switch {
	case assetTexture.Flags.Has(dto.TextureFlag2D):
		return LoadTexture2D(loader, assetTexture)
	case assetTexture.Flags.Has(dto.TextureFlagCubeMap):
		return LoadTextureCube(loader, assetTexture)
	default:
		return Identifiable[render.Texture]{}, fmt.Errorf("unsupported texture type (flags: %v)", assetTexture.Flags)
	}
}

// LoadTexture2D resolves a 2D texture from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadTexture2D(loader *AssetLoader, assetTexture dto.Texture) (Identifiable[render.Texture], error) {
	var texture render.Texture

	allocateTexture := func() error {
		renderAPI := loader.Engine().Graphics().API()
		texture = renderAPI.CreateColorTexture2D(render.ColorTexture2DInfo{
			GenerateMipmaps: assetTexture.Flags.Has(dto.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(dto.TextureFlagLinearSpace),
			Format:          resolveDataFormat(assetTexture.Format),
			MipmapLayers: gog.Map(assetTexture.MipmapLayers, func(layer dto.MipmapLayer) render.Mipmap2DLayer {
				return render.Mipmap2DLayer{
					Width:  layer.Width,
					Height: layer.Height,
					Data:   layer.Layers[0].Data,
				}
			}),
		})
		return nil
	}
	if err := loader.ScheduleMain(allocateTexture).Wait(); err != nil {
		return Identifiable[render.Texture]{}, err
	}

	return Identifiable[render.Texture]{
		ID:    assetTexture.ID,
		Value: texture,
	}, nil
}

// LoadTextureCube resolves a cube map texture from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadTextureCube(loader *AssetLoader, assetTexture dto.Texture) (Identifiable[render.Texture], error) {
	var texture render.Texture

	allocateTexture := func() error {
		renderAPI := loader.Engine().Graphics().API()
		texture = renderAPI.CreateColorTextureCube(render.ColorTextureCubeInfo{
			GenerateMipmaps: assetTexture.Flags.Has(dto.TextureFlagMipmapping),
			GammaCorrection: !assetTexture.Flags.Has(dto.TextureFlagLinearSpace),
			Format:          resolveDataFormat(assetTexture.Format),
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
	}
	if err := loader.ScheduleMain(allocateTexture).Wait(); err != nil {
		return Identifiable[render.Texture]{}, err
	}

	return Identifiable[render.Texture]{
		ID:    assetTexture.ID,
		Value: texture,
	}, nil
}

// LoadTextures resolves a list of textures from the given asset textures.
//
// This is a blocking operation and should be called from a worker thread.
func LoadTextures(loader *AssetLoader, assetTextures []dto.Texture) (IdentifiableList[render.Texture], error) {
	textures := make(IdentifiableList[render.Texture], len(assetTextures))
	var group errgroup.Group
	for i, assetTexture := range assetTextures {
		group.Go(func() error {
			texture, err := LoadTexture(loader, assetTexture)
			textures[i] = texture
			return err
		})
	}
	return textures, group.Wait()
}

// UnloadTexture unloads a texture from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadTexture(loader *AssetLoader, idTexture Identifiable[render.Texture]) error {
	texture := idTexture.Value
	return loader.ScheduleMain(func() error {
		texture.Release()
		return nil
	}).Wait()
}

// UnloadTextures unloads a list of textures from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadTextures(loader *AssetLoader, idTextures IdentifiableList[render.Texture]) error {
	var group errgroup.Group
	for _, idTexture := range idTextures {
		group.Go(func() error {
			return UnloadTexture(loader, idTexture)
		})
	}
	return group.Wait()
}

// LoadMaterial resolves a material from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadMaterial(loader *AssetLoader, assetMaterial dto.Material, shaders IdentifiableList[*graphics.Shader], textures IdentifiableList[render.Texture]) (Identifiable[*graphics.Material], error) {
	geometryPasses, err := LoadMaterialPasses(loader, assetMaterial.GeometryPasses, shaders)
	if err != nil {
		return Identifiable[*graphics.Material]{}, fmt.Errorf("failed to convert geometry passes: %w", err)
	}
	shadowPasses, err := LoadMaterialPasses(loader, assetMaterial.ShadowPasses, shaders)
	if err != nil {
		return Identifiable[*graphics.Material]{}, fmt.Errorf("failed to convert shadow passes: %w", err)
	}
	forwardPasses, err := LoadMaterialPasses(loader, assetMaterial.ForwardPasses, shaders)
	if err != nil {
		return Identifiable[*graphics.Material]{}, fmt.Errorf("failed to convert forward passes: %w", err)
	}
	skyPasses, err := LoadMaterialPasses(loader, assetMaterial.SkyPasses, shaders)
	if err != nil {
		return Identifiable[*graphics.Material]{}, fmt.Errorf("failed to convert sky passes: %w", err)
	}
	postprocessingPasses, err := LoadMaterialPasses(loader, assetMaterial.PostprocessingPasses, shaders)
	if err != nil {
		return Identifiable[*graphics.Material]{}, fmt.Errorf("failed to convert postprocessing passes: %w", err)
	}
	materialInfo := graphics.MaterialInfo{
		Name:                 assetMaterial.Name,
		GeometryPasses:       geometryPasses,
		ShadowPasses:         shadowPasses,
		ForwardPasses:        forwardPasses,
		SkyPasses:            skyPasses,
		PostprocessingPasses: postprocessingPasses,
	}

	var material *graphics.Material
	allocateMaterial := func() error {
		gfxEngine := loader.Engine().Graphics()
		renderAPI := gfxEngine.API()
		material = gfxEngine.CreateMaterial(materialInfo)
		for _, binding := range assetMaterial.Textures {
			texture, ok := textures.FindByID(binding.TextureID)
			if !ok {
				return fmt.Errorf("texture with ID %d not found", binding.TextureID)
			}
			material.SetTexture(binding.BindingName, texture)
		}
		for _, binding := range assetMaterial.Textures {
			sampler := renderAPI.CreateSampler(render.SamplerInfo{
				Wrapping:   resolveWrapMode(binding.Wrapping),
				Filtering:  resolveFiltering(binding.Filtering),
				Mipmapping: binding.Mipmapping,
			})
			material.SetSampler(binding.BindingName, sampler)
		}
		for _, binding := range assetMaterial.Properties {
			material.SetProperty(binding.BindingName, binding.Data)
		}
		return nil
	}
	if err := loader.ScheduleMain(allocateMaterial).Wait(); err != nil {
		return Identifiable[*graphics.Material]{}, err
	}

	return Identifiable[*graphics.Material]{
		ID:    assetMaterial.ID,
		Value: material,
	}, nil
}

// LoadMaterialPass resolves a material pass from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadMaterialPass(loader *AssetLoader, assetPass dto.MaterialPass, shaders IdentifiableList[*graphics.Shader]) (graphics.MaterialPassInfo, error) {
	shader, ok := shaders.FindByID(assetPass.ShaderID)
	if !ok {
		return graphics.MaterialPassInfo{}, fmt.Errorf("shader with ID %d not found", assetPass.ShaderID)
	}
	return graphics.MaterialPassInfo{
		Layer:           assetPass.Layer,
		Culling:         opt.V(resolveCullMode(assetPass.Culling)),
		FrontFace:       opt.V(resolveFaceOrientation(assetPass.FrontFace)),
		DepthTest:       opt.V(assetPass.DepthTest),
		DepthWrite:      opt.V(assetPass.DepthWrite),
		DepthComparison: opt.V(resolveComparison(assetPass.DepthComparison)),
		Blending:        opt.V(assetPass.Blending),
		Shader:          shader,
	}, nil
}

// LoadMaterialPasses resolves a list of material passes from the given asset
// data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadMaterialPasses(loader *AssetLoader, assetPasses []dto.MaterialPass, shaders IdentifiableList[*graphics.Shader]) ([]graphics.MaterialPassInfo, error) {
	result := make([]graphics.MaterialPassInfo, len(assetPasses))
	for i, assetPass := range assetPasses {
		passInfo, err := LoadMaterialPass(loader, assetPass, shaders)
		if err != nil {
			return nil, err
		}
		result[i] = passInfo
	}
	return result, nil
}

// LoadMaterials resolves a list of materials from the given asset
// materials.
//
// This is a blocking operation and should be called from a worker thread.
func LoadMaterials(loader *AssetLoader, assetMaterials []dto.Material, shaders IdentifiableList[*graphics.Shader], textures IdentifiableList[render.Texture]) (IdentifiableList[*graphics.Material], error) {
	materials := make(IdentifiableList[*graphics.Material], len(assetMaterials))
	var group errgroup.Group
	for i, assetMaterial := range assetMaterials {
		group.Go(func() error {
			material, err := LoadMaterial(loader, assetMaterial, shaders, textures)
			materials[i] = material
			return err
		})
	}
	return materials, group.Wait()
}

// UnloadMaterial unloads a material from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadMaterial(loader *AssetLoader, idMaterial Identifiable[*graphics.Material]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadMaterials unloads a list of materials from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadMaterials(loader *AssetLoader, idMaterials IdentifiableList[*graphics.Material]) error {
	for _, idMaterial := range idMaterials {
		if err := UnloadMaterial(loader, idMaterial); err != nil {
			return err
		}
	}
	return nil
}

func resolveShaderType(assetType dto.ShaderType) graphics.ShaderType {
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

func resolveDataFormat(format dto.TexelFormat) render.DataFormat {
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

func resolveWrapMode(wrap dto.WrapMode) render.WrapMode {
	switch wrap {
	case dto.WrapModeClamp:
		return render.WrapModeClamp
	case dto.WrapModeRepeat:
		return render.WrapModeRepeat
	case dto.WrapModeMirroredRepeat:
		return render.WrapModeMirroredRepeat
	default:
		panic(fmt.Errorf("unknown wrap mode: %v", wrap))
	}
}

func resolveFiltering(filter dto.FilterMode) render.FilterMode {
	switch filter {
	case dto.FilterModeNearest:
		return render.FilterModeNearest
	case dto.FilterModeLinear:
		return render.FilterModeLinear
	case dto.FilterModeAnisotropic:
		return render.FilterModeAnisotropic
	default:
		panic(fmt.Errorf("unknown filter mode: %v", filter))
	}
}

func resolveCullMode(mode dto.CullMode) render.CullMode {
	switch mode {
	case dto.CullModeNone:
		return render.CullModeNone
	case dto.CullModeFront:
		return render.CullModeFront
	case dto.CullModeBack:
		return render.CullModeBack
	case dto.CullModeFrontAndBack:
		return render.CullModeFrontAndBack
	default:
		panic(fmt.Errorf("unknown cull mode: %v", mode))
	}
}

func resolveFaceOrientation(orientation dto.FaceOrientation) render.FaceOrientation {
	switch orientation {
	case dto.FaceOrientationCCW:
		return render.FaceOrientationCCW
	case dto.FaceOrientationCW:
		return render.FaceOrientationCW
	default:
		panic(fmt.Errorf("unknown face orientation: %v", orientation))
	}
}

func resolveComparison(comparison dto.Comparison) render.Comparison {
	switch comparison {
	case dto.ComparisonNever:
		return render.ComparisonNever
	case dto.ComparisonLess:
		return render.ComparisonLess
	case dto.ComparisonEqual:
		return render.ComparisonEqual
	case dto.ComparisonLessOrEqual:
		return render.ComparisonLessOrEqual
	case dto.ComparisonGreater:
		return render.ComparisonGreater
	case dto.ComparisonNotEqual:
		return render.ComparisonNotEqual
	case dto.ComparisonGreaterOrEqual:
		return render.ComparisonGreaterOrEqual
	case dto.ComparisonAlways:
		return render.ComparisonAlways
	default:
		panic(fmt.Errorf("unknown comparison: %v", comparison))
	}
}
