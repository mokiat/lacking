package pack

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/physics"
)

type SaveLevelAssetAction struct {
	resource      asset.Resource
	levelProvider LevelProvider
}

func (a *SaveLevelAssetAction) Describe() string {
	return fmt.Sprintf("save_level_asset(%q)", a.resource.Name())
}

func (a *SaveLevelAssetAction) Run() error {
	level := a.levelProvider.Level()

	conv := newConverter(false)

	modelAsset := asset.Model{}

	bodyDefinitions := make([]asset.BodyDefinition, len(level.CollisionMeshes))
	bodyInstances := make([]asset.BodyInstance, len(level.CollisionMeshes))
	for i, collisionMesh := range level.CollisionMeshes {
		collisionMeshAsset := asset.CollisionMesh{
			Triangles:   make([]asset.CollisionTriangle, len(collisionMesh.Triangles)),
			Translation: dprec.ZeroVec3(),
			Rotation:    dprec.IdentityQuat(),
		}
		for j, triangle := range collisionMesh.Triangles {
			collisionMeshAsset.Triangles[j] = asset.CollisionTriangle{
				A: stod.Vec3(triangle.A),
				B: stod.Vec3(triangle.B),
				C: stod.Vec3(triangle.C),
			}
		}
		bodyDefinitions[i] = asset.BodyDefinition{
			Name:                   fmt.Sprintf("static-%d", i),
			Mass:                   1.0,
			MomentOfInertia:        physics.SymmetricMomentOfInertia(1.0),
			RestitutionCoefficient: 1.0,
			DragFactor:             0.0,
			AngularDragFactor:      0.0,
			CollisionMeshes: []asset.CollisionMesh{
				collisionMeshAsset,
			},
		}
		bodyInstances[i] = asset.BodyInstance{
			NodeIndex: -1,
			BodyIndex: int32(i),
		}
	}
	modelAsset.BodyDefinitions = bodyDefinitions
	modelAsset.BodyInstances = bodyInstances

	materials := make([]asset.Material, len(level.Materials))
	for i, staticMaterial := range level.Materials {
		materials[i] = conv.BuildMaterial(staticMaterial)
		conv.assetMaterialIndexFromMaterial[staticMaterial] = i
	}
	modelAsset.Materials = materials

	meshDefinitions := make([]asset.MeshDefinition, len(level.StaticMeshes))
	meshInstances := make([]asset.MeshInstance, len(level.StaticMeshes))
	for i, staticMesh := range level.StaticMeshes {
		meshDefinitions[i] = conv.BuildMeshDefinition(staticMesh)
		meshInstances[i] = asset.MeshInstance{
			Name:            staticMesh.Name,
			NodeIndex:       -1,
			ArmatureIndex:   -1,
			DefinitionIndex: int32(i),
		}
	}
	modelAsset.MeshDefinitions = meshDefinitions
	modelAsset.MeshInstances = meshInstances

	modelInstances := make([]asset.ModelInstance, len(level.StaticEntities))
	for i, staticEntity := range level.StaticEntities {
		t, r, s := stod.Mat4(staticEntity.Matrix).TRS()
		modelInstances[i] = asset.ModelInstance{
			ModelIndex:  -1,
			ModelID:     staticEntity.Model,
			Name:        staticEntity.Name,
			Translation: t,
			Rotation:    r,
			Scale:       s,
		}
	}

	levelAsset := &asset.Scene{
		ModelInstances: modelInstances,
	}
	if err := a.resource.WriteContent(levelAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}
