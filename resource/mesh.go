package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/game/graphics"
)

type Mesh struct {
	Name string

	GFXMeshTemplate graphics.MeshTemplate

	materials []graphics.Material
	textures  []*TwoDTexture
}

func AllocateMesh(registry *Registry, name string, gfxWorker *async.Worker, gfxEngine graphics.Engine, meshAsset *asset.Mesh) (*Mesh, error) {
	mesh := &Mesh{
		Name: name,
	}

	subMeshDefinitions := make([]graphics.SubMeshTemplateDefinition, len(meshAsset.SubMeshes))
	for i, subMeshAsset := range meshAsset.SubMeshes {
		var metalnessTexture graphics.TwoDTexture
		if subMeshAsset.Material.MetalnessTexture != "" {
			var texture *TwoDTexture
			result := registry.LoadTwoDTexture(subMeshAsset.Material.MetalnessTexture).
				OnSuccess(InjectTwoDTexture(&texture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load metalness texture: %w", err)
			}
			metalnessTexture = texture.GFXTexture
			mesh.textures = append(mesh.textures, texture)
		}

		var roughnessTexture graphics.TwoDTexture
		if subMeshAsset.Material.RoughnessTexture != "" {
			var texture *TwoDTexture
			result := registry.LoadTwoDTexture(subMeshAsset.Material.RoughnessTexture).
				OnSuccess(InjectTwoDTexture(&texture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load roughness texture: %w", err)
			}
			roughnessTexture = texture.GFXTexture
			mesh.textures = append(mesh.textures, texture)
		}

		var albedoTexture graphics.TwoDTexture
		if subMeshAsset.Material.ColorTexture != "" {
			var texture *TwoDTexture
			result := registry.LoadTwoDTexture(subMeshAsset.Material.ColorTexture).
				OnSuccess(InjectTwoDTexture(&texture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load albedo texture: %w", err)
			}
			albedoTexture = texture.GFXTexture
			mesh.textures = append(mesh.textures, texture)
		}

		var normalTexture graphics.TwoDTexture
		if subMeshAsset.Material.NormalTexture != "" {
			var texture *TwoDTexture
			result := registry.LoadTwoDTexture(subMeshAsset.Material.NormalTexture).
				OnSuccess(InjectTwoDTexture(&texture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load normal texture: %w", err)
			}
			normalTexture = texture.GFXTexture
			mesh.textures = append(mesh.textures, texture)
		}

		var material graphics.Material
		gfxWorker.Schedule(async.VoidTask(func() error {
			definition := graphics.PBRMaterialDefinition{
				BackfaceCulling:  subMeshAsset.Material.BackfaceCulling,
				Metalness:        subMeshAsset.Material.Metalness,
				MetalnessTexture: metalnessTexture,
				Roughness:        subMeshAsset.Material.Roughness,
				RoughnessTexture: roughnessTexture,
				AlbedoColor: sprec.NewVec4(
					subMeshAsset.Material.Color[0],
					subMeshAsset.Material.Color[1],
					subMeshAsset.Material.Color[2],
					subMeshAsset.Material.Color[3],
				),
				AlbedoTexture: albedoTexture,
				NormalScale:   subMeshAsset.Material.NormalScale,
				NormalTexture: normalTexture,
			}
			material = gfxEngine.CreatePBRMaterial(definition)
			return nil
		})).Wait()
		mesh.materials = append(mesh.materials, material)

		subMeshDefinition := graphics.SubMeshTemplateDefinition{
			Primitive:   assetToGraphicsPrimitive(subMeshAsset.Primitive),
			IndexOffset: int(subMeshAsset.IndexOffset),
			IndexCount:  int(subMeshAsset.IndexCount),
			Material:    material,
		}
		subMeshDefinitions[i] = subMeshDefinition
	}

	gfxWorker.Schedule(async.VoidTask(func() error {
		definition := graphics.MeshTemplateDefinition{
			VertexData: meshAsset.VertexData,
			VertexFormat: graphics.VertexFormat{
				HasCoord:            meshAsset.VertexLayout.CoordOffset != asset.UnspecifiedOffset,
				CoordOffsetBytes:    int(meshAsset.VertexLayout.CoordOffset),
				CoordStrideBytes:    int(meshAsset.VertexLayout.CoordStride),
				HasNormal:           meshAsset.VertexLayout.NormalOffset != asset.UnspecifiedOffset,
				NormalOffsetBytes:   int(meshAsset.VertexLayout.NormalOffset),
				NormalStrideBytes:   int(meshAsset.VertexLayout.NormalStride),
				HasTangent:          meshAsset.VertexLayout.TangentOffset != asset.UnspecifiedOffset,
				TangentOffsetBytes:  int(meshAsset.VertexLayout.TangentOffset),
				TangentStrideBytes:  int(meshAsset.VertexLayout.TangentStride),
				HasTexCoord:         meshAsset.VertexLayout.TexCoordOffset != asset.UnspecifiedOffset,
				TexCoordOffsetBytes: int(meshAsset.VertexLayout.TexCoordOffset),
				TexCoordStrideBytes: int(meshAsset.VertexLayout.TexCoordStride),
				HasColor:            meshAsset.VertexLayout.ColorOffset != asset.UnspecifiedOffset,
				ColorOffsetBytes:    int(meshAsset.VertexLayout.ColorOffset),
				ColorStrideBytes:    int(meshAsset.VertexLayout.ColorStride),
			},
			IndexData:   meshAsset.IndexData,
			IndexFormat: graphics.IndexFormatU16,
			SubMeshes:   subMeshDefinitions,
		}
		mesh.GFXMeshTemplate = gfxEngine.CreateMeshTemplate(definition)
		return nil
	})).Wait()

	return mesh, nil
}

func ReleaseMesh(registry *Registry, gfxWorker *async.Worker, mesh *Mesh) error {
	gfxWorker.Schedule(async.VoidTask(func() error {
		mesh.GFXMeshTemplate.Delete()
		return nil
	})).Wait()

	for _, material := range mesh.materials {
		gfxWorker.Schedule(async.VoidTask(func() error {
			material.Delete()
			return nil
		})).Wait()
	}

	for _, texture := range mesh.textures {
		if result := registry.UnloadTwoDTexture(texture).Wait(); result.Err != nil {
			return result.Err
		}
	}

	mesh.GFXMeshTemplate = nil
	mesh.materials = nil
	mesh.textures = nil
	return nil
}

func assetToGraphicsPrimitive(primitive asset.Primitive) graphics.Primitive {
	switch primitive {
	case asset.PrimitivePoints:
		return graphics.PrimitivePoints
	case asset.PrimitiveLines:
		return graphics.PrimitiveLines
	case asset.PrimitiveLineStrip:
		return graphics.PrimitiveLineStrip
	case asset.PrimitiveLineLoop:
		return graphics.PrimitiveLineStrip
	case asset.PrimitiveTriangles:
		return graphics.PrimitiveTriangles
	case asset.PrimitiveTriangleStrip:
		return graphics.PrimitiveTriangleStrip
	case asset.PrimitiveTriangleFan:
		return graphics.PrimitiveTriangleFan
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}
