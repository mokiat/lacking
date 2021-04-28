package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
)

type SaveLevelAssetAction struct {
	locator       AssetLocator
	uri           string
	levelProvider LevelProvider
}

func (a *SaveLevelAssetAction) Describe() string {
	return fmt.Sprintf("save_level_asset(uri: %q)", a.uri)
}

func (a *SaveLevelAssetAction) Run() error {
	level := a.levelProvider.Level()

	levelAsset := &asset.Level{
		SkyboxTexture:            level.SkyboxTexture,
		AmbientReflectionTexture: level.AmbientReflectionTexture,
		AmbientRefractionTexture: level.AmbientRefractionTexture,
		CollisionMeshes:          make([]asset.LevelCollisionMesh, len(level.CollisionMeshes)),
		StaticMeshes:             make([]asset.Mesh, len(level.StaticMeshes)),
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

	for i, staticMesh := range level.StaticMeshes {
		levelAsset.StaticMeshes[i] = meshToAssetMesh(&staticMesh)
	}

	for i, staticEntity := range level.StaticEntities {
		staticEntityAsset := asset.LevelEntity{
			Model:  staticEntity.Model,
			Matrix: staticEntity.Matrix.ColumnMajorArray(),
		}
		levelAsset.StaticEntities[i] = staticEntityAsset
	}

	out, err := a.locator.Create(a.uri)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := asset.EncodeLevel(out, levelAsset); err != nil {
		return fmt.Errorf("failed to encode asset: %w", err)
	}
	return nil
}
