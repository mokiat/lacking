package physics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/placement3d"
)

type TerrainID struct {
	index    int32
	revision int32
}

var NilTerrainID = TerrainID{}

type TerrainView struct {
	scene *Scene
}

// TODO: Rework the Terrain such that the Mesh is attached similar to how
// collision shapes are attached to bodies. This should open the door for
// other types of terrain, such as heightmaps, etc.
//
// However, this requires rework of placement3d API.

func (v TerrainView) Create(position dprec.Vec3, rotation dprec.Quat, mesh CollisionMesh) TerrainID {
	index, terrain := v.scene.allocateTerrain()

	meshID := v.scene.collisionScene.CreateMesh(placement3d.MeshInfo[terrainData]{
		Position:  opt.V(position),
		Rotation:  opt.V(rotation),
		Mesh:      mesh.Shape,
		Filtering: mesh.Filtering,
		UserData: terrainData{
			index:                  index,
			frictionCoefficient:    mesh.FrictionCoefficient,
			restitutionCoefficient: mesh.RestitutionCoefficient,
		},
	})

	*terrain = terrainState{
		meshID:   meshID,
		revision: terrain.revision + 1, // progress revision to valid (odd) value
	}

	return TerrainID{
		index:    index,
		revision: terrain.revision,
	}
}

func (v TerrainView) Delete(id TerrainID) {
	terrain := v.resolve(id, true)

	v.scene.collisionScene.DeleteMesh(terrain.meshID)

	*terrain = terrainState{
		meshID:   placement3d.InvalidMeshID,
		revision: terrain.revision + 1, // progress revision to invalid (even) value
	}

	v.scene.releaseTerrain(id.index)
}

func (v TerrainView) Each(cb func(id TerrainID)) {
	v.scene.eachTerrain(func(index int, terrain *terrainState) {
		cb(TerrainID{
			index:    int32(index),
			revision: terrain.revision,
		})
	})
}

func (v TerrainView) Handle(id TerrainID) TerrainHandle {
	return TerrainHandle{
		view: v,
		id:   id,
	}
}

func (v TerrainView) IsValid(id TerrainID) bool {
	terrain := v.resolve(id, false)
	return terrain != nil
}

func (v TerrainView) resolve(id TerrainID, required bool) *terrainState {
	if id.revision == 0 {
		if required {
			panic("invalid terrain ID")
		}
		return nil
	}
	terrain := &v.scene.terrains[id.index]
	if terrain.revision != id.revision {
		if required {
			panic("invalid terrain ID")
		}
		return nil
	}
	return terrain
}

type TerrainHandle struct {
	view TerrainView
	id   TerrainID
}

type terrainState struct {
	meshID   placement3d.MeshID
	revision int32
}

func (s *terrainState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
