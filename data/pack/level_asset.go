package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
	gameasset "github.com/mokiat/lacking/game/asset"
)

type SaveLevelAssetAction struct {
	registry      gameasset.Registry
	id            string
	levelProvider LevelProvider
}

func (a *SaveLevelAssetAction) Describe() string {
	return fmt.Sprintf("save_level_asset(id: %q)", a.id)
}

func (a *SaveLevelAssetAction) Run() error {
	level := a.levelProvider.Level()

	conv := newConverter(false)

	levelAsset := &asset.Level{
		SkyboxTexture:            level.SkyboxTexture,
		AmbientReflectionTexture: level.AmbientReflectionTexture,
		AmbientRefractionTexture: level.AmbientRefractionTexture,
		Materials:                make([]gameasset.Material, len(level.Materials)),
		CollisionMeshes:          make([]asset.LevelCollisionMesh, len(level.CollisionMeshes)),
		StaticMeshes:             make([]gameasset.MeshDefinition, len(level.StaticMeshes)),
		StaticEntities:           make([]asset.LevelEntity, len(level.StaticEntities)),
	}

	for i, collisionMesh := range level.CollisionMeshes {
		collisionMeshAsset := asset.LevelCollisionMesh{
			Triangles: make([]asset.Triangle, len(collisionMesh.Triangles)),
		}
		for j, triangle := range collisionMesh.Triangles {
			collisionMeshAsset.Triangles[j] = asset.Triangle{
				asset.Point{triangle.A.X, triangle.A.Y, triangle.A.Z},
				asset.Point{triangle.B.X, triangle.B.Y, triangle.B.Z},
				asset.Point{triangle.C.X, triangle.C.Y, triangle.C.Z},
			}
		}
		levelAsset.CollisionMeshes[i] = collisionMeshAsset
	}

	for i, staticMaterial := range level.Materials {
		levelAsset.Materials[i] = conv.BuildMaterial(staticMaterial)
		conv.assetMaterialIndexFromMaterial[staticMaterial] = i
	}

	for i, staticMesh := range level.StaticMeshes {
		levelAsset.StaticMeshes[i] = conv.BuildMeshDefinition(staticMesh)
	}

	for i, staticEntity := range level.StaticEntities {
		staticEntityAsset := asset.LevelEntity{
			Model:  staticEntity.Model,
			Matrix: staticEntity.Matrix.ColumnMajorArray(),
		}
		levelAsset.StaticEntities[i] = staticEntityAsset
	}

	resource := a.registry.ResourceByID(a.id)
	if resource == nil {
		resource = a.registry.CreateIDResource(a.id, "level", a.id)
	}
	if err := resource.WriteContent(levelAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	if err := a.registry.Save(); err != nil {
		return fmt.Errorf("error saving resources: %w", err)
	}
	return nil
}
