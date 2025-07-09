package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/util/shape3d"
	"golang.org/x/sync/errgroup"
)

// LoadPhysicsMaterial loads a physics material from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsMaterial(loader *AssetLoader, assetMaterial dto.BodyMaterial) (Identifiable[*physics.Material], error) {
	materialInfo := physics.MaterialInfo{
		FrictionCoefficient:    assetMaterial.FrictionCoefficient,
		RestitutionCoefficient: assetMaterial.RestitutionCoefficient,
	}

	var material *physics.Material
	allocateMaterial := func() error {
		physicsEngine := loader.Engine().Physics()
		material = physicsEngine.CreateMaterial(materialInfo)
		return nil
	}
	if err := loader.ScheduleMain(allocateMaterial).Wait(); err != nil {
		return Identifiable[*physics.Material]{}, err
	}

	return Identifiable[*physics.Material]{
		ID:    assetMaterial.ID,
		Value: material,
	}, nil
}

// LoadPhysicsMaterials loads a list of physics materials from the given asset
// materials.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsMaterials(loader *AssetLoader, assetMaterials []dto.BodyMaterial) (IdentifiableList[*physics.Material], error) {
	materials := make(IdentifiableList[*physics.Material], len(assetMaterials))
	var group errgroup.Group
	for i, assetMaterial := range assetMaterials {
		group.Go(func() error {
			material, err := LoadPhysicsMaterial(loader, assetMaterial)
			materials[i] = material
			return err
		})
	}
	return materials, group.Wait()
}

// UnloadPhysicsMaterial unloads a physics material from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPhysicsMaterial(loader *AssetLoader, idMaterial Identifiable[*physics.Material]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadPhysicsMaterials unloads a list of physics materials from the asset
// loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPhysicsMaterials(loader *AssetLoader, idMaterials IdentifiableList[*physics.Material]) error {
	for _, idMaterial := range idMaterials {
		if err := UnloadPhysicsMaterial(loader, idMaterial); err != nil {
			return err
		}
	}
	return nil
}

// LoadPhysicsBodyDefinition loads a physics body definition from the given
// asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsBodyDefinition(loader *AssetLoader, assetBodyDefinition dto.BodyDefinition, materials IdentifiableList[*physics.Material]) (Identifiable[*physics.BodyDefinition], error) {
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
		CollisionSpheres:       resolveCollisionSpheres(assetBodyDefinition),
		CollisionBoxes:         resolveCollisionBoxes(assetBodyDefinition),
		CollisionMeshes:        resolveCollisionMeshes(assetBodyDefinition),
	}

	var bodyDefinition *physics.BodyDefinition
	allocateDefinition := func() error {
		physicsEngine := loader.Engine().Physics()
		bodyDefinition = physicsEngine.CreateBodyDefinition(bodyDefinitionInfo)
		return nil
	}
	if err := loader.ScheduleMain(allocateDefinition).Wait(); err != nil {
		return Identifiable[*physics.BodyDefinition]{}, err
	}

	return Identifiable[*physics.BodyDefinition]{
		ID:    assetBodyDefinition.ID,
		Value: bodyDefinition,
	}, nil
}

// LoadPhysicsBodyDefinitions loads a list of physics body definitions from the
// given asset body definitions.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsBodyDefinitions(loader *AssetLoader, assetBodyDefinitions []dto.BodyDefinition, materials IdentifiableList[*physics.Material]) (IdentifiableList[*physics.BodyDefinition], error) {
	bodyDefinitions := make(IdentifiableList[*physics.BodyDefinition], len(assetBodyDefinitions))
	var group errgroup.Group
	for i, assetBodyDefinition := range assetBodyDefinitions {
		group.Go(func() error {
			bodyDefinition, err := LoadPhysicsBodyDefinition(loader, assetBodyDefinition, materials)
			bodyDefinitions[i] = bodyDefinition
			return err
		})
	}
	return bodyDefinitions, group.Wait()
}

// UnloadPhysicsBodyDefinition unloads a physics body definition from the asset
// loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPhysicsBodyDefinition(loader *AssetLoader, idBodyDefinition Identifiable[*physics.BodyDefinition]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadPhysicsBodyDefinitions unloads a list of physics body definitions from
// the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPhysicsBodyDefinitions(loader *AssetLoader, idBodyDefinitions IdentifiableList[*physics.BodyDefinition]) error {
	for _, idBodyDefinition := range idBodyDefinitions {
		if err := UnloadPhysicsBodyDefinition(loader, idBodyDefinition); err != nil {
			return err
		}
	}
	return nil
}

func resolveCollisionSpheres(bodyDef dto.BodyDefinition) []shape3d.Sphere {
	result := make([]shape3d.Sphere, len(bodyDef.CollisionSpheres))
	for i, collisionSphereAsset := range bodyDef.CollisionSpheres {
		result[i] = shape3d.NewSphere(
			collisionSphereAsset.Translation,
			collisionSphereAsset.Radius,
		)
	}
	return result
}

func resolveCollisionBoxes(bodyDef dto.BodyDefinition) []shape3d.Box {
	result := make([]shape3d.Box, len(bodyDef.CollisionBoxes))
	for i, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result[i] = shape3d.NewBox(
			collisionBoxAsset.Translation,
			collisionBoxAsset.Rotation,
			dprec.NewVec3(collisionBoxAsset.Width, collisionBoxAsset.Height, collisionBoxAsset.Length),
		)
	}
	return result
}

func resolveCollisionMeshes(bodyDef dto.BodyDefinition) []shape3d.Mesh {
	result := make([]shape3d.Mesh, len(bodyDef.CollisionMeshes))
	for i, collisionMeshAsset := range bodyDef.CollisionMeshes {
		transform := shape3d.TRTransform(collisionMeshAsset.Translation, collisionMeshAsset.Rotation)
		triangles := make([]shape3d.Triangle, len(collisionMeshAsset.Triangles))
		for j, triangleAsset := range collisionMeshAsset.Triangles {
			triangles[j] = shape3d.NewTriangle(
				transform.Apply(triangleAsset.A),
				transform.Apply(triangleAsset.B),
				transform.Apply(triangleAsset.C),
			)
		}
		result[i] = shape3d.NewMesh(triangles)
	}
	return result
}

// BodyTemplate represents a template for physics body that can be
// instantiated in a scene.
type BodyTemplate struct {
	NodeID     uint32
	Definition *physics.BodyDefinition
}

// LoadPhysicsBodyTemplate resolves a physics body template from the given asset
// data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsBodyTemplate(loader *AssetLoader, assetBody dto.Body, bodyDefinitions IdentifiableList[*physics.BodyDefinition]) (Identifiable[BodyTemplate], error) {
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

// LoadPhysicsBodyTemplates resolves a list of physics body templates from the
// given asset bodies.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPhysicsBodyTemplates(loader *AssetLoader, assetBodies []dto.Body, bodyDefinitions IdentifiableList[*physics.BodyDefinition]) (IdentifiableList[BodyTemplate], error) {
	bodyTemplates := make(IdentifiableList[BodyTemplate], len(assetBodies))
	for i, assetBody := range assetBodies {
		template, err := LoadPhysicsBodyTemplate(loader, assetBody, bodyDefinitions)
		if err != nil {
			return IdentifiableList[BodyTemplate]{}, err
		}
		bodyTemplates[i] = template
	}
	return bodyTemplates, nil
}

// UnloadPhysicsBodyTemplate unloads a physics body template from the asset
// loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPhysicsBodyTemplate(loader *AssetLoader, idBody Identifiable[BodyTemplate]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadPhysicsBodyTemplates unloads a list of physics body templates from the
// asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPhysicsBodyTemplates(loader *AssetLoader, idBodies IdentifiableList[BodyTemplate]) error {
	for _, idBody := range idBodies {
		if err := UnloadPhysicsBodyTemplate(loader, idBody); err != nil {
			return err
		}
	}
	return nil
}

// InstantiatePhysicsBodyTemplateStatic creates a static physics body in the
// given scene from the provided body template.
//
// This operation needs to be called from the main thread.
func InstantiatePhysicsBodyTemplateStatic(scene *Scene, template BodyTemplate, nodes IdentifiableList[*hierarchy.Node]) {
	node := nodes.GetByID(template.NodeID)
	absMatrix := node.AbsoluteMatrix()
	transform := shape3d.TRTransform(absMatrix.Translation(), absMatrix.Rotation())

	scene.physicsScene.CreateProp(physics.PropInfo{
		Name: node.Name(),
		CollisionSpheres: gog.Map(template.Definition.CollisionSpheres(), func(sphere shape3d.Sphere) shape3d.Sphere {
			return shape3d.TransformedSphere(sphere, transform)
		}),
		CollisionBoxes: gog.Map(template.Definition.CollisionBoxes(), func(box shape3d.Box) shape3d.Box {
			return shape3d.TransformedBox(box, transform)
		}),
		CollisionMeshes: gog.Map(template.Definition.CollisionMeshes(), func(mesh shape3d.Mesh) shape3d.Mesh {
			return shape3d.TransformedMesh(mesh, transform)
		}),
	})
}

// InstantiatePhysicsBodyTemplateDynamic creates a dynamic physics body in the
// given scene from the provided body template and returns it.
//
// This operation needs to be called from the main thread.
func InstantiatePhysicsBodyTemplateDynamic(scene *Scene, template BodyTemplate, nodes IdentifiableList[*hierarchy.Node]) physics.Body {
	node := nodes.GetByID(template.NodeID)
	translation, rotation, _ := node.AbsoluteMatrix().TRS()
	body := scene.physicsScene.CreateBody(physics.BodyInfo{
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
