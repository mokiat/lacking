package conv

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
)

type PhysicsSource interface {
	AllPhysicsBodyPlacements() []mdl.Placed[*mdl.PhysicsBody]
	AllPhysicsTerrainPlacements() []mdl.Placed[*mdl.PhysicsTerrain]
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
	allBodyPlacements := src.AllPhysicsBodyPlacements()

	dtoBodies := make([]dto.PhysicsBody, len(allBodyPlacements))
	for i, placement := range allBodyPlacements {
		body := placement.Value

		var dtoCollisionSpheres []dto.CollisionSphere
		for _, sphere := range body.CollisionSpheres() {
			dtoCollisionSpheres = append(dtoCollisionSpheres, dto.CollisionSphere{
				CollisionShape: dto.CollisionShape{
					FrictionCoefficient:    sphere.FrictionCoefficient(),
					RestitutionCoefficient: sphere.RestitutionCoefficient(),
				},
				Translation: sphere.Translation(),
				Radius:      sphere.Radius(),
			})
		}

		var dtoCollisionBoxes []dto.CollisionBox
		for _, box := range body.CollisionBoxes() {
			dtoCollisionBoxes = append(dtoCollisionBoxes, dto.CollisionBox{
				CollisionShape: dto.CollisionShape{
					FrictionCoefficient:    box.FrictionCoefficient(),
					RestitutionCoefficient: box.RestitutionCoefficient(),
				},
				Translation: box.Translation(),
				Rotation:    box.Rotation(),
				Width:       box.Width(),
				Height:      box.Height(),
				Length:      box.Length(),
			})
		}

		dtoBodies[i] = dto.PhysicsBody{
			ID:               body.ID(),
			NodeID:           placement.Node.ID(),
			Mass:             body.Mass(),
			MomentOfInertia:  body.MomentOfInertia(),
			CollisionSpheres: dtoCollisionSpheres,
			CollisionBoxes:   dtoCollisionBoxes,
		}
	}

	allTerrainPlacements := src.AllPhysicsTerrainPlacements()
	dtoTerrains := make([]dto.PhysicsTerrain, len(allTerrainPlacements))
	for i, placement := range allTerrainPlacements {
		terrain := placement.Value

		var dtoCollisionMeshes []dto.CollisionMesh
		for _, mesh := range terrain.CollisionMeshes() {
			dtoCollisionMeshes = append(dtoCollisionMeshes, dto.CollisionMesh{
				CollisionShape: dto.CollisionShape{
					FrictionCoefficient:    mesh.FrictionCoefficient(),
					RestitutionCoefficient: mesh.RestitutionCoefficient(),
				},
				Translation: mesh.Translation(),
				Rotation:    mesh.Rotation(),
				Triangles: gog.Map(mesh.Triangles(), func(triangle mdl.CollisionTriangle) dto.CollisionTriangle {
					return dto.CollisionTriangle{
						A: triangle.A,
						B: triangle.B,
						C: triangle.C,
					}
				}),
			})
		}

		dtoTerrains[i] = dto.PhysicsTerrain{
			ID:              terrain.ID(),
			NodeID:          placement.Node.ID(),
			CollisionMeshes: dtoCollisionMeshes,
		}
	}

	return &dto.PhysicsChunk{
		Bodies:   dtoBodies,
		Terrains: dtoTerrains,
	}, nil
}
