package resource

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
)

type Material struct {
	Name string

	GFXMaterial *graphics.Material
	textures    []*TwoDTexture
}

func AllocateMaterial(registry *Registry, gfxEngine *graphics.Engine, materialAsset *asset.Material) (*Material, error) {
	assetPBR := asset.NewPBRMaterialView(materialAsset)

	material := &Material{
		Name: materialAsset.Name,
	}

	var albedoTexture *graphics.TwoDTexture
	if assetPBR.BaseColorTexture() != "" {
		var texture *TwoDTexture
		result := registry.LoadTwoDTexture(assetPBR.BaseColorTexture()).
			OnSuccess(InjectTwoDTexture(&texture)).
			Wait()
		if err := result.Err; err != nil {
			return nil, fmt.Errorf("failed to load albedo texture: %w", err)
		}
		albedoTexture = texture.GFXTexture
		material.textures = append(material.textures, texture)
	}

	var metallicRoughnessTexture *graphics.TwoDTexture
	if assetPBR.MetallicRoughnessTexture() != "" {
		var texture *TwoDTexture
		result := registry.LoadTwoDTexture(assetPBR.MetallicRoughnessTexture()).
			OnSuccess(InjectTwoDTexture(&texture)).
			Wait()
		if err := result.Err; err != nil {
			return nil, fmt.Errorf("failed to load metallicRoughness texture: %w", err)
		}
		metallicRoughnessTexture = texture.GFXTexture
		material.textures = append(material.textures, texture)
	}

	var normalTexture *graphics.TwoDTexture
	if assetPBR.NormalTexture() != "" {
		var texture *TwoDTexture
		result := registry.LoadTwoDTexture(assetPBR.NormalTexture()).
			OnSuccess(InjectTwoDTexture(&texture)).
			Wait()
		if err := result.Err; err != nil {
			return nil, fmt.Errorf("failed to load normal texture: %w", err)
		}
		normalTexture = texture.GFXTexture
		material.textures = append(material.textures, texture)
	}

	registry.ScheduleVoid(func() {
		definition := graphics.PBRMaterialDefinition{
			BackfaceCulling:          materialAsset.BackfaceCulling,
			AlphaBlending:            materialAsset.Blending,
			AlphaTesting:             materialAsset.AlphaTesting,
			AlphaThreshold:           materialAsset.AlphaThreshold,
			Metallic:                 assetPBR.Metallic(),
			Roughness:                assetPBR.Roughness(),
			MetallicRoughnessTexture: metallicRoughnessTexture,
			AlbedoColor:              assetPBR.BaseColor(),
			AlbedoTexture:            albedoTexture,
			NormalScale:              assetPBR.NormalScale(),
			NormalTexture:            normalTexture,
		}
		material.GFXMaterial = gfxEngine.CreatePBRMaterial(definition)
	}).Wait()

	return material, nil
}

func ReleaseMaterial(registry *Registry, material *Material) error {
	registry.ScheduleVoid(func() {
		material.GFXMaterial.Delete()
	}).Wait()

	for _, texture := range material.textures {
		if result := registry.UnloadTwoDTexture(texture).Wait(); result.Err != nil {
			return result.Err
		}
	}

	material.GFXMaterial = nil
	material.textures = nil
	return nil
}
