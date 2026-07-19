package game

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/spatial/shape3d"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
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
		material = physics.NewMaterial(materialInfo)
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
		bodyDefinition = physics.NewBodyDefinition(bodyDefinitionInfo)
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
		result[i] = shape3d.Sphere{
			Center: collisionSphereAsset.Translation,
			Radius: collisionSphereAsset.Radius,
		}
	}
	return result
}

func resolveCollisionBoxes(bodyDef dto.BodyDefinition) []shape3d.Box {
	result := make([]shape3d.Box, len(bodyDef.CollisionBoxes))
	for i, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result[i] = shape3d.Box{
			Center:     collisionBoxAsset.Translation,
			Rotation:   shape3d.RotationFromQuat(collisionBoxAsset.Rotation),
			HalfWidth:  collisionBoxAsset.Width / 2.0,
			HalfHeight: collisionBoxAsset.Height / 2.0,
			HalfLength: collisionBoxAsset.Length / 2.0,
		}
	}
	return result
}

func resolveCollisionMeshes(bodyDef dto.BodyDefinition) []shape3d.Mesh {
	result := make([]shape3d.Mesh, len(bodyDef.CollisionMeshes))
	for i, collisionMeshAsset := range bodyDef.CollisionMeshes {
		transform := shape3d.TRTransform(
			collisionMeshAsset.Translation,
			shape3d.RotationFromQuat(collisionMeshAsset.Rotation),
		)
		triangles := make([]shape3d.Triangle, len(collisionMeshAsset.Triangles))
		for j, triangleAsset := range collisionMeshAsset.Triangles {
			triangles[j] = shape3d.Triangle{
				A: transform.Apply(triangleAsset.A),
				B: transform.Apply(triangleAsset.B),
				C: transform.Apply(triangleAsset.C),
			}
		}
		result[i] = shape3d.Mesh{
			Triangles: triangles,
		}
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
func InstantiatePhysicsBodyTemplateStatic(scene *Scene, template BodyTemplate, nodes IdentifiableList[hierarchy.NodeID]) {
	node := nodes.GetByID(template.NodeID)
	absMatrix := scene.Hierarchy().NodeAbsoluteMatrix(node)
	nodeName := scene.Hierarchy().NodeName(node)
	scene.physicsScene.CreateProp(physics.PropInfo{
		Name:             nodeName,
		Position:         opt.V(absMatrix.Translation()),
		Rotation:         opt.V(absMatrix.Rotation()),
		CollisionSpheres: template.Definition.CollisionSpheres(),
		CollisionBoxes:   template.Definition.CollisionBoxes(),
		CollisionMeshes:  template.Definition.CollisionMeshes(),
	})
}

// InstantiatePhysicsBodyTemplateDynamic creates a dynamic physics body in the
// given scene from the provided body template and returns it.
//
// This operation needs to be called from the main thread.
func InstantiatePhysicsBodyTemplateDynamic(scene *Scene, template BodyTemplate, nodes IdentifiableList[hierarchy.NodeID]) physics.Body {
	node := nodes.GetByID(template.NodeID)
	absMatrix := scene.Hierarchy().NodeAbsoluteMatrix(node)
	nodeName := scene.Hierarchy().NodeName(node)
	translation, rotation, _ := absMatrix.TRS()
	body := scene.physicsScene.CreateBody(physics.BodyInfo{
		Name:       nodeName,
		Definition: template.Definition,
		Position:   translation,
		Rotation:   rotation,
	})
	scene.bodyBindingSet.Bind(node, body)
	return body
}
