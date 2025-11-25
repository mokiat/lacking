package conv

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
)

type PhysicsSource interface {
	AllPhysicsBodyMaterials() []*mdl.BodyMaterial
	AllPhysicsBodyDefinitions() []*mdl.BodyDefinition
	AllPhysicsBodyPlacements() []mdl.Placed[*mdl.Body]
}

func NewPhysicsConverter() *PhysicsConverter {
	return &PhysicsConverter{}
}

type PhysicsConverter struct{}

func (c *PhysicsConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(PhysicsSource)
	if !ok {
		return nil
	}
	chunk, err := c.CreatePhysicsChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(dto.PhysicsChunkID, chunk))
	return nil
}

func (c *PhysicsConverter) CreatePhysicsChunk(src PhysicsSource) (*dto.PhysicsChunk, error) {
	allMaterials := src.AllPhysicsBodyMaterials()
	dtoBodyMaterials := make([]dto.BodyMaterial, len(allMaterials))
	for i, material := range allMaterials {
		dtoBodyMaterials[i] = c.convertBodyMaterial(material)
	}

	allDefinitions := src.AllPhysicsBodyDefinitions()
	dtoBodyDefinitions := make([]dto.BodyDefinition, len(allDefinitions))
	for i, definition := range allDefinitions {
		dtoBodyDefinitions[i] = c.convertBodyDefinition(definition)
	}

	allBodyPlacements := src.AllPhysicsBodyPlacements()
	dtoBodies := make([]dto.Body, len(allBodyPlacements))
	for i, placement := range allBodyPlacements {
		body := placement.Value
		dtoBodies[i] = c.convertBody(placement.Node, body)
	}

	return &dto.PhysicsChunk{
		BodyMaterials:   dtoBodyMaterials,
		BodyDefinitions: dtoBodyDefinitions,
		Bodies:          dtoBodies,
	}, nil
}

func (c *PhysicsConverter) convertBodyMaterial(material *mdl.BodyMaterial) dto.BodyMaterial {
	return dto.BodyMaterial{
		ID:                     material.ID(),
		FrictionCoefficient:    material.FrictionCoefficient(),
		RestitutionCoefficient: material.RestitutionCoefficient(),
	}
}

func (c *PhysicsConverter) convertBodyDefinition(definition *mdl.BodyDefinition) dto.BodyDefinition {
	return dto.BodyDefinition{
		ID:                definition.ID(),
		MaterialID:        definition.Material().ID(),
		Mass:              definition.Mass(),
		MomentOfInertia:   definition.MomentOfInertia(),
		DragFactor:        definition.DragFactor(),
		AngularDragFactor: definition.AngularDragFactor(),
		CollisionBoxes: gog.Map(definition.CollisionBoxes(), func(box *mdl.CollisionBox) dto.CollisionBox {
			return dto.CollisionBox{
				Translation: box.Translation(),
				Rotation:    box.Rotation(),
				Width:       box.Width(),
				Height:      box.Height(),
				Length:      box.Length(),
			}
		}),
		CollisionSpheres: gog.Map(definition.CollisionSpheres(), func(sphere *mdl.CollisionSphere) dto.CollisionSphere {
			return dto.CollisionSphere{
				Translation: sphere.Translation(),
				Radius:      sphere.Radius(),
			}
		}),
		CollisionMeshes: gog.Map(definition.CollisionMeshes(), func(mesh *mdl.CollisionMesh) dto.CollisionMesh {
			return dto.CollisionMesh{
				Translation: mesh.Translation(),
				Rotation:    mesh.Rotation(),
				Triangles: gog.Map(mesh.Triangles(), func(triangle mdl.CollisionTriangle) dto.CollisionTriangle {
					return dto.CollisionTriangle{
						A: triangle.A,
						B: triangle.B,
						C: triangle.C,
					}
				}),
			}
		}),
	}
}

func (c *PhysicsConverter) convertBody(node *mdl.Node, body *mdl.Body) dto.Body {
	return dto.Body{
		ID:               body.ID(),
		NodeID:           node.ID(),
		BodyDefinitionID: body.Definition().ID(),
	}
}
