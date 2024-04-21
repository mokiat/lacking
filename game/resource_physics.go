package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertBodyMaterial(assetBodyMaterial asset.BodyMaterial) async.Promise[*physics.Material] {
	materialInfo := physics.MaterialInfo{
		FrictionCoefficient:    assetBodyMaterial.FrictionCoefficient,
		RestitutionCoefficient: assetBodyMaterial.RestitutionCoefficient,
	}
	promise := async.NewPromise[*physics.Material]()
	s.gfxWorker.Schedule(func() {
		physicsEngine := s.engine.Physics()
		material := physicsEngine.CreateMaterial(materialInfo)
		promise.Deliver(material)
	})
	return promise
}

func (s *ResourceSet) convertBodyDefinition(bodyMaterials []*physics.Material, assetBodyDefinition asset.BodyDefinition) async.Promise[*physics.BodyDefinition] {
	material := bodyMaterials[assetBodyDefinition.MaterialIndex]

	bodyDefinitionInfo := physics.BodyDefinitionInfo{
		Mass:                   assetBodyDefinition.Mass,
		MomentOfInertia:        assetBodyDefinition.MomentOfInertia,
		FrictionCoefficient:    material.FrictionCoefficient(),
		RestitutionCoefficient: material.RestitutionCoefficient(),
		DragFactor:             assetBodyDefinition.DragFactor,
		AngularDragFactor:      assetBodyDefinition.AngularDragFactor,
		AerodynamicShapes:      nil, // TODO
		CollisionSpheres:       s.convertCollisionSpheres(assetBodyDefinition),
		CollisionBoxes:         s.convertCollisionBoxes(assetBodyDefinition),
		CollisionMeshes:        s.convertCollisionMeshes(assetBodyDefinition),
	}

	promise := async.NewPromise[*physics.BodyDefinition]()
	s.gfxWorker.Schedule(func() {
		physicsEngine := s.engine.Physics()
		bodyDefinition := physicsEngine.CreateBodyDefinition(bodyDefinitionInfo)
		promise.Deliver(bodyDefinition)
	})
	return promise
}

func (s *ResourceSet) convertCollisionSpheres(bodyDef asset.BodyDefinition) []collision.Sphere {
	result := make([]collision.Sphere, len(bodyDef.CollisionSpheres))
	for i, collisionSphereAsset := range bodyDef.CollisionSpheres {
		result[i] = collision.NewSphere(
			collisionSphereAsset.Translation,
			collisionSphereAsset.Radius,
		)
	}
	return result
}

func (s *ResourceSet) convertCollisionBoxes(bodyDef asset.BodyDefinition) []collision.Box {
	result := make([]collision.Box, len(bodyDef.CollisionBoxes))
	for i, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result[i] = collision.NewBox(
			collisionBoxAsset.Translation,
			collisionBoxAsset.Rotation,
			dprec.NewVec3(collisionBoxAsset.Width, collisionBoxAsset.Height, collisionBoxAsset.Lenght),
		)
	}
	return result
}

func (s *ResourceSet) convertCollisionMeshes(bodyDef asset.BodyDefinition) []collision.Mesh {
	result := make([]collision.Mesh, len(bodyDef.CollisionMeshes))
	for i, collisionMeshAsset := range bodyDef.CollisionMeshes {
		transform := collision.TRTransform(collisionMeshAsset.Translation, collisionMeshAsset.Rotation)
		triangles := make([]collision.Triangle, len(collisionMeshAsset.Triangles))
		for j, triangleAsset := range collisionMeshAsset.Triangles {
			template := collision.NewTriangle(
				triangleAsset.A,
				triangleAsset.B,
				triangleAsset.C,
			)
			triangles[j].Replace(template, transform)
		}
		result[i] = collision.NewMesh(triangles)
	}
	return result
}

func (s *ResourceSet) convertBody(assetBody asset.Body) bodyInstance {
	return bodyInstance{
		NodeIndex:       int(assetBody.NodeIndex),
		DefinitionIndex: int(assetBody.BodyDefinitionIndex),
	}
}
