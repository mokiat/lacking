package physicsconv

import (
	"iter"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/asset/dto/physicsdto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

type Source interface {
	NodesIter() iter.Seq2[int, *mdl.Node]
}

func CreatePhysicsChunk(src Source) *physicsdto.PhysicsChunk {
	ctx := &conversionContext{
		convertedBodyMaterials:   make(map[*mdl.BodyMaterial]uint32),
		convertedBodyDefinitions: make(map[*mdl.BodyDefinition]uint32),
	}

	var dtoBodies []physicsdto.Body
	for i, node := range src.NodesIter() {
		switch source := node.Source().(type) {
		case *mdl.Body:
			dtoBody := convertBody(ctx, uint32(i), source)
			dtoBodies = append(dtoBodies, dtoBody)
		}
	}

	return &physicsdto.PhysicsChunk{
		BodyMaterials:   ctx.dtoBodyMaterials,
		BodyDefinitions: ctx.dtoBodyDefinitions,
		Bodies:          dtoBodies,
	}
}

func convertBody(ctx *conversionContext, nodeIndex uint32, body *mdl.Body) physicsdto.Body {
	bodyDefinitionIndex := convertBodyDefinition(ctx, body.Definition())
	return physicsdto.Body{
		NodeIndex:           nodeIndex,
		BodyDefinitionIndex: bodyDefinitionIndex,
	}
}

func convertBodyDefinition(ctx *conversionContext, definition *mdl.BodyDefinition) uint32 {
	if index, ok := ctx.convertedBodyDefinitions[definition]; ok {
		return index
	}

	materialIndex := convertBodyMaterial(ctx, definition.Material())

	assetDefinition := physicsdto.BodyDefinition{
		MaterialIndex:     materialIndex,
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

	index := uint32(len(ctx.dtoBodyDefinitions))
	ctx.dtoBodyDefinitions = append(ctx.dtoBodyDefinitions, assetDefinition)
	ctx.convertedBodyDefinitions[definition] = index
	return index
}

func convertBodyMaterial(ctx *conversionContext, material *mdl.BodyMaterial) uint32 {
	if index, ok := ctx.convertedBodyMaterials[material]; ok {
		return index
	}

	assetMaterial := physicsdto.BodyMaterial{
		FrictionCoefficient:    material.FrictionCoefficient(),
		RestitutionCoefficient: material.RestitutionCoefficient(),
	}

	index := uint32(len(ctx.dtoBodyMaterials))
	ctx.dtoBodyMaterials = append(ctx.dtoBodyMaterials, assetMaterial)
	ctx.convertedBodyMaterials[material] = index
	return index
}

type conversionContext struct {
	dtoBodyMaterials       []physicsdto.BodyMaterial
	convertedBodyMaterials map[*mdl.BodyMaterial]uint32

	dtoBodyDefinitions       []physicsdto.BodyDefinition
	convertedBodyDefinitions map[*mdl.BodyDefinition]uint32
}
