package physicsconv

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/asset/dto/physicsdto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

type Source interface {
	AllPhysicsBodyMaterials() []*mdl.BodyMaterial
	AllPhysicsBodyDefinitions() []*mdl.BodyDefinition
	AllPhysicsBodyPlacements() []mdl.Placed[*mdl.Body]
}

func CreatePhysicsChunk(src Source) (*physicsdto.PhysicsChunk, error) {
	allMaterials := src.AllPhysicsBodyMaterials()
	dtoBodyMaterials := make([]physicsdto.BodyMaterial, len(allMaterials))
	for i, material := range allMaterials {
		dtoBodyMaterials[i] = convertBodyMaterial(material)
	}

	allDefinitions := src.AllPhysicsBodyDefinitions()
	dtoBodyDefinitions := make([]physicsdto.BodyDefinition, len(allDefinitions))
	for i, definition := range allDefinitions {
		dtoBodyDefinitions[i] = convertBodyDefinition(definition)
	}

	allBodyPlacements := src.AllPhysicsBodyPlacements()
	dtoBodies := make([]physicsdto.Body, len(allBodyPlacements))
	for i, placement := range allBodyPlacements {
		body := placement.Value
		dtoBodies[i] = convertBody(placement.Node, body)
	}

	return &physicsdto.PhysicsChunk{
		BodyMaterials:   dtoBodyMaterials,
		BodyDefinitions: dtoBodyDefinitions,
		Bodies:          dtoBodies,
	}, nil
}

func convertBodyMaterial(material *mdl.BodyMaterial) physicsdto.BodyMaterial {
	return physicsdto.BodyMaterial{
		ID:                     material.ID(),
		FrictionCoefficient:    material.FrictionCoefficient(),
		RestitutionCoefficient: material.RestitutionCoefficient(),
	}
}

func convertBodyDefinition(definition *mdl.BodyDefinition) physicsdto.BodyDefinition {
	return physicsdto.BodyDefinition{
		ID:                definition.ID(),
		MaterialID:        definition.Material().ID(),
		Mass:              definition.Mass(),
		MomentOfInertia:   definition.MomentOfInertia(),
		DragFactor:        definition.DragFactor(),
		AngularDragFactor: definition.AngularDragFactor(),
		CollisionBoxes: gog.Map(definition.CollisionBoxes(), func(box *mdl.CollisionBox) physicsdto.CollisionBox {
			return physicsdto.CollisionBox{
				Translation: box.Translation(),
				Rotation:    box.Rotation(),
				Width:       box.Width(),
				Height:      box.Height(),
				Length:      box.Length(),
			}
		}),
		CollisionSpheres: gog.Map(definition.CollisionSpheres(), func(sphere *mdl.CollisionSphere) physicsdto.CollisionSphere {
			return physicsdto.CollisionSphere{
				Translation: sphere.Translation(),
				Radius:      sphere.Radius(),
			}
		}),
		CollisionMeshes: gog.Map(definition.CollisionMeshes(), func(mesh *mdl.CollisionMesh) physicsdto.CollisionMesh {
			return physicsdto.CollisionMesh{
				Translation: mesh.Translation(),
				Rotation:    mesh.Rotation(),
				Triangles: gog.Map(mesh.Triangles(), func(triangle mdl.CollisionTriangle) physicsdto.CollisionTriangle {
					return physicsdto.CollisionTriangle{
						A: triangle.A,
						B: triangle.B,
						C: triangle.C,
					}
				}),
			}
		}),
	}
}

func convertBody(node *mdl.Node, body *mdl.Body) physicsdto.Body {
	return physicsdto.Body{
		ID:               body.ID(),
		NodeID:           node.ID(),
		BodyDefinitionID: body.Definition().ID(),
	}
}
