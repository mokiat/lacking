package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
)

type Material struct {
	Name string

	GFXMaterial *graphics.Material
	textures    []*TwoDTexture
}

func AllocateMaterial(registry *Registry, gfxEngine *graphics.Engine, materialAsset *asset.Material) (*Material, error) {
	material := &Material{
		Name: materialAsset.Name,
	}

	var albedoTexture *graphics.TwoDTexture
	if materialAsset.Textures[0] != "" {
		var texture *TwoDTexture
		result := registry.LoadTwoDTexture(materialAsset.Textures[0]).
			OnSuccess(InjectTwoDTexture(&texture)).
			Wait()
		if err := result.Err; err != nil {
			return nil, fmt.Errorf("failed to load albedo texture: %w", err)
		}
		albedoTexture = texture.GFXTexture
		material.textures = append(material.textures, texture)
	}

	var metallicRoughnessTexture *graphics.TwoDTexture
	if materialAsset.Textures[1] != "" {
		var texture *TwoDTexture
		result := registry.LoadTwoDTexture(materialAsset.Textures[1]).
			OnSuccess(InjectTwoDTexture(&texture)).
			Wait()
		if err := result.Err; err != nil {
			return nil, fmt.Errorf("failed to load metallicRoughness texture: %w", err)
		}
		metallicRoughnessTexture = texture.GFXTexture
		material.textures = append(material.textures, texture)
	}

	var normalTexture *graphics.TwoDTexture
	if materialAsset.Textures[2] != "" {
		var texture *TwoDTexture
		result := registry.LoadTwoDTexture(materialAsset.Textures[2]).
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
			Metallic:                 materialAsset.Scalars[4], // FIXME: potential nil pointer deref
			Roughness:                materialAsset.Scalars[5], // FIXME: potential nil pointer deref
			MetallicRoughnessTexture: metallicRoughnessTexture,
			AlbedoColor: sprec.NewVec4(
				materialAsset.Scalars[0], // FIXME: potential nil pointer deref
				materialAsset.Scalars[1], // FIXME: potential nil pointer deref
				materialAsset.Scalars[2], // FIXME: potential nil pointer deref
				materialAsset.Scalars[3], // FIXME: potential nil pointer deref
			),
			AlbedoTexture: albedoTexture,
			NormalScale:   materialAsset.Scalars[6], // FIXME: potential nil pointer deref
			NormalTexture: normalTexture,
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
