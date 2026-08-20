package game

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/physics"
)

type TerrainTemplate struct {
	NodeID uint32

	CollisionMeshes []physics.CollisionMesh
}

// LoadPhysicsTerrainTemplate resolves a physics body template from the given asset
// data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsTerrainTemplate(loader *AssetLoader, assetTerrain dto.PhysicsTerrain) (Identifiable[TerrainTemplate], error) {
	return Identifiable[TerrainTemplate]{
		ID: assetTerrain.ID,
		Value: TerrainTemplate{
			NodeID: assetTerrain.NodeID,
			CollisionMeshes: gog.Map(assetTerrain.CollisionMeshes, func(collisionMesh dto.CollisionMesh) physics.CollisionMesh {
				transform := shape3d.TRTransform(
					collisionMesh.Translation,
					shape3d.RotationFromQuat(collisionMesh.Rotation),
				)
				triangles := gog.Map(collisionMesh.Triangles, func(triangle dto.CollisionTriangle) shape3d.Triangle {
					return shape3d.Triangle{
						A: transform.Apply(triangle.A),
						B: transform.Apply(triangle.B),
						C: transform.Apply(triangle.C),
					}
				})
				return physics.CollisionMesh{
					Shape:                  shape3d.NewMesh(triangles),
					FrictionCoefficient:    collisionMesh.FrictionCoefficient,
					RestitutionCoefficient: collisionMesh.RestitutionCoefficient,
					Filtering:              placement3d.FilterInfo{},
				}
			}),
		},
	}, nil
}

// LoadPhysicsTerrainTemplates resolves a list of physics body templates from the
// given asset bodies.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsTerrainTemplates(loader *AssetLoader, assetTerrains []dto.PhysicsTerrain) (IdentifiableList[TerrainTemplate], error) {
	terrainTemplates := make(IdentifiableList[TerrainTemplate], len(assetTerrains))
	for i, assetTerrain := range assetTerrains {
		template, err := LoadPhysicsTerrainTemplate(loader, assetTerrain)
		if err != nil {
			return IdentifiableList[TerrainTemplate]{}, err
		}
		terrainTemplates[i] = template
	}
	return terrainTemplates, nil
}
