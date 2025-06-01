package game

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"golang.org/x/sync/errgroup"
)

func (l *AssetLoader) ResolvePhysicsMaterial(assetMaterial dto.BodyMaterial) (Identifiable[*physics.Material], error) {
	materialInfo := physics.MaterialInfo{
		FrictionCoefficient:    assetMaterial.FrictionCoefficient,
		RestitutionCoefficient: assetMaterial.RestitutionCoefficient,
	}

	var material *physics.Material
	allocateMaterial := func(engine *Engine) error {
		physicsEngine := engine.Physics()
		material = physicsEngine.CreateMaterial(materialInfo)
		return nil
	}
	if err := l.ScheduleMain(allocateMaterial).Wait(); err != nil {
		return Identifiable[*physics.Material]{}, err
	}

	return Identifiable[*physics.Material]{
		ID:    assetMaterial.ID,
		Value: material,
	}, nil
}

func (l *AssetLoader) ResolvePhysicsMaterials(assetMaterials []dto.BodyMaterial) (IdentifiableList[*physics.Material], error) {
	materials := make(IdentifiableList[*physics.Material], len(assetMaterials))
	var group errgroup.Group
	for i, assetMaterial := range assetMaterials {
		group.Go(func() error {
			material, err := l.ResolvePhysicsMaterial(assetMaterial)
			materials[i] = material
			return err
		})
	}
	return materials, group.Wait()
}

func (l *AssetLoader) ResolvePhysicsBodyDefinition(assetBodyDefinition dto.BodyDefinition, materials IdentifiableList[*physics.Material]) (Identifiable[*physics.BodyDefinition], error) {
	material, ok := materials.FindByID(assetBodyDefinition.MaterialID)
	if !ok {
		return Identifiable[*physics.BodyDefinition]{}, fmt.Errorf("physics material with ID %d not found", assetBodyDefinition.MaterialID)
	}

	bodyDefinitionInfo := physics.BodyDefinitionInfo{
		Mass:                   assetBodyDefinition.Mass,
		MomentOfInertia:        assetBodyDefinition.MomentOfInertia,
		FrictionCoefficient:    material.FrictionCoefficient(),
		RestitutionCoefficient: material.RestitutionCoefficient(),
		DragFactor:             assetBodyDefinition.DragFactor,
		AngularDragFactor:      assetBodyDefinition.AngularDragFactor,
		AerodynamicShapes:      nil, // TODO
		CollisionSpheres:       l.resolveCollisionSpheres(assetBodyDefinition),
		CollisionBoxes:         l.resolveCollisionBoxes(assetBodyDefinition),
		CollisionMeshes:        l.resolveCollisionMeshes(assetBodyDefinition),
	}

	var bodyDefinition *physics.BodyDefinition
	allocateDefinition := func(engine *Engine) error {
		physicsEngine := engine.Physics()
		bodyDefinition = physicsEngine.CreateBodyDefinition(bodyDefinitionInfo)
		return nil
	}
	if err := l.ScheduleMain(allocateDefinition).Wait(); err != nil {
		return Identifiable[*physics.BodyDefinition]{}, err
	}

	return Identifiable[*physics.BodyDefinition]{
		ID:    assetBodyDefinition.ID,
		Value: bodyDefinition,
	}, nil
}

func (l *AssetLoader) ResolvePhysicsBodyDefinitions(assetBodyDefinitions []dto.BodyDefinition, materials IdentifiableList[*physics.Material]) (IdentifiableList[*physics.BodyDefinition], error) {
	bodyDefinitions := make(IdentifiableList[*physics.BodyDefinition], len(assetBodyDefinitions))
	var group errgroup.Group
	for i, assetBodyDefinition := range assetBodyDefinitions {
		group.Go(func() error {
			bodyDefinition, err := l.ResolvePhysicsBodyDefinition(assetBodyDefinition, materials)
			bodyDefinitions[i] = bodyDefinition
			return err
		})
	}
	return bodyDefinitions, group.Wait()
}

func (l *AssetLoader) resolveCollisionSpheres(bodyDef dto.BodyDefinition) []collision.Sphere {
	result := make([]collision.Sphere, len(bodyDef.CollisionSpheres))
	for i, collisionSphereAsset := range bodyDef.CollisionSpheres {
		result[i] = collision.NewSphere(
			collisionSphereAsset.Translation,
			collisionSphereAsset.Radius,
		)
	}
	return result
}

func (l *AssetLoader) resolveCollisionBoxes(bodyDef dto.BodyDefinition) []collision.Box {
	result := make([]collision.Box, len(bodyDef.CollisionBoxes))
	for i, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result[i] = collision.NewBox(
			collisionBoxAsset.Translation,
			collisionBoxAsset.Rotation,
			dprec.NewVec3(collisionBoxAsset.Width, collisionBoxAsset.Height, collisionBoxAsset.Length),
		)
	}
	return result
}

func (l *AssetLoader) resolveCollisionMeshes(bodyDef dto.BodyDefinition) []collision.Mesh {
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

type BodyTemplate struct {
	NodeID     uint32
	Definition *physics.BodyDefinition
}

func (l *AssetLoader) ResolvePhysicsBodyTemplate(assetBody dto.Body, bodyDefinitions IdentifiableList[*physics.BodyDefinition]) (Identifiable[BodyTemplate], error) {
	bodyDefinition, ok := bodyDefinitions.FindByID(assetBody.BodyDefinitionID)
	if !ok {
		return Identifiable[BodyTemplate]{}, fmt.Errorf("body definition with ID %d not found", assetBody.BodyDefinitionID)
	}
	return Identifiable[BodyTemplate]{
		ID: assetBody.ID,
		Value: BodyTemplate{
			NodeID:     assetBody.NodeID,
			Definition: bodyDefinition,
		},
	}, nil
}

func (l *AssetLoader) ResolvePhysicsBodyTemplates(assetBodies []dto.Body, bodyDefinitions IdentifiableList[*physics.BodyDefinition]) (IdentifiableList[BodyTemplate], error) {
	bodyTemplates := make(IdentifiableList[BodyTemplate], len(assetBodies))
	for i, assetBody := range assetBodies {
		template, err := l.ResolvePhysicsBodyTemplate(assetBody, bodyDefinitions)
		if err != nil {
			return IdentifiableList[BodyTemplate]{}, err
		}
		bodyTemplates[i] = template
	}
	return bodyTemplates, nil
}

func (s *Scene) InstantiatePhysicsBodyTemplateStatic(template BodyTemplate, nodes IdentifiableList[*hierarchy.Node]) {
	node, ok := nodes.FindByID(template.NodeID)
	if !ok {
		return
	}
	absMatrix := node.AbsoluteMatrix()
	transform := collision.TRTransform(absMatrix.Translation(), absMatrix.Rotation())
	collisionSet := collision.NewSet()
	collisionSet.Replace(template.Definition.CollisionSet(), transform)
	s.physicsScene.CreateProp(physics.PropInfo{
		Name:         node.Name(),
		CollisionSet: collisionSet,
	})
}

func (s *Scene) InstantiatePhysicsBodyTemplateDynamic(template BodyTemplate, nodes IdentifiableList[*hierarchy.Node]) physics.Body {
	node := nodes.GetByID(template.NodeID)
	translation, rotation, _ := node.AbsoluteMatrix().TRS()
	body := s.physicsScene.CreateBody(physics.BodyInfo{
		Name:       node.Name(),
		Definition: template.Definition,
		Position:   translation,
		Rotation:   rotation,
	})
	node.SetSource(BodyNodeSource{
		Body: body,
	})
	return body
}
