package physics

import (
	"github.com/mokiat/lacking/core/spatial/placement3d"
)

// TerrainID uniquely identifies a terrain that was created through
// [TerrainView.Create].
//
// The zero value, also available as [NilTerrainID], does not reference a
// valid terrain.
type TerrainID struct {
	index    int32
	revision int32
}

// NilTerrainID is a [TerrainID] that never references a valid terrain.
var NilTerrainID = TerrainID{}

// TerrainView provides access to the terrains that belong to a [Scene].
//
// A terrain is static, immovable geometry that bodies collide with. Unlike
// a body, it has no mass, velocity or transform of its own.
//
// A TerrainView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Terrains].
type TerrainView struct {
	scene *Scene
}

// TODO: Rework the Terrain such that the Mesh is attached similar to how
// collision shapes are attached to bodies. This should open the door for
// other types of terrain, such as heightmaps, etc.
//
// However, this requires rework of placement3d API.

// Create allocates a new terrain, initially without any collision shapes,
// and returns its ID.
func (v TerrainView) Create() TerrainID {
	index, terrain := v.scene.allocateTerrain()

	terrainID := v.scene.collisionScene.CreateTerrain(placement3d.TerrainInfo[terrainData]{
		UserData: terrainData{
			index: index,
		},
	})

	*terrain = terrainState{
		terrainID: terrainID,
		revision:  terrain.revision + 1, // progress revision to valid (odd) value
	}

	return TerrainID{
		index:    index,
		revision: terrain.revision,
	}
}

// CreateHandle behaves like [TerrainView.Create] but wraps the resulting
// ID in a [TerrainHandle], as returned by [TerrainView.Handle], for
// callers that want to keep acting on the new terrain without holding
// onto its ID separately.
func (v TerrainView) CreateHandle() TerrainHandle {
	return v.Handle(v.Create())
}

// Delete removes the terrain with the specified ID, together with its
// collision shapes.
//
// It panics if the ID does not reference a valid terrain. Use
// [TerrainView.IsValid] first if the ID's validity is not otherwise
// guaranteed.
func (v TerrainView) Delete(id TerrainID) {
	terrain := v.resolve(id, true)

	v.scene.collisionScene.DeleteTerrain(terrain.terrainID)

	*terrain = terrainState{
		terrainID: placement3d.NilTerrainID,
		revision:  terrain.revision + 1, // progress revision to invalid (even) value
	}

	v.scene.releaseTerrain(id.index)
}

// Each calls cb once for every terrain that is currently alive within
// this Scene, in unspecified order.
func (v TerrainView) Each(cb func(id TerrainID)) {
	v.scene.eachTerrain(func(index int, terrain *terrainState) {
		cb(TerrainID{
			index:    int32(index),
			revision: terrain.revision,
		})
	})
}

// Handle returns a [TerrainHandle] that wraps the specified ID, as a more
// convenient means of repeatedly accessing the same terrain without
// having to pass its ID to this view on every call.
func (v TerrainView) Handle(id TerrainID) TerrainHandle {
	return TerrainHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references a terrain that has
// not been deleted.
func (v TerrainView) IsValid(id TerrainID) bool {
	terrain := v.resolve(id, false)
	return terrain != nil
}

// AttachCollisionMesh attaches a mesh collision shape to the specified
// terrain and returns an ID through which it can later be detached. The
// mesh is specified in world space, since a terrain has no transform of
// its own.
//
// It panics if the ID does not reference a valid terrain, or if the mesh
// is empty.
func (v TerrainView) AttachCollisionMesh(id TerrainID, col CollisionMesh) TerrainCollisionShapeID {
	terrain := v.resolve(id, true)
	shapeID := v.scene.collisionScene.AttachMesh(terrain.terrainID, placement3d.MeshInfo[shapeData]{
		Mesh:      col.Shape,
		Filtering: col.Filtering,
		UserData: shapeData{
			frictionCoefficient:    col.FrictionCoefficient,
			restitutionCoefficient: col.RestitutionCoefficient,
		},
	})
	return TerrainCollisionShapeID{
		terrainID: id,
		shapeID:   shapeID,
	}
}

// DetachCollisionShape removes a collision shape that was previously
// attached to the specified terrain through
// [TerrainView.AttachCollisionMesh].
//
// It panics if shapeID was not obtained for this same terrain.
func (v TerrainView) DetachCollisionShape(id TerrainID, shapeID TerrainCollisionShapeID) {
	if id != shapeID.terrainID {
		panic("invalid shape ID for terrain")
	}
	v.scene.collisionScene.DeleteTerrainShape(shapeID.shapeID)
}

func (v TerrainView) idFromIndex(index int32) TerrainID {
	terrain := &v.scene.terrains[index]
	return TerrainID{
		index:    index,
		revision: terrain.revision,
	}
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

// TerrainHandle is an object-oriented alternative to [TerrainView] that
// is bound to a specific [TerrainID].
//
// It is obtained through [TerrainView.Handle].
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [TerrainID] (which holds no pointers), a Handle keeps an internal
// reference to the owning [Scene]. This keeps that Scene reachable for as
// long as the handle is retained, and adds a pointer that the garbage
// collector must trace wherever the handle is stored. Prefer storing IDs
// over handles in long-lived collections unless the convenience is worth
// that cost.
type TerrainHandle struct {
	view TerrainView
	id   TerrainID
}

// ID returns the [TerrainID] wrapped by this handle.
func (v TerrainHandle) ID() TerrainID {
	return v.id
}

// Delete removes the wrapped terrain.
//
// See [TerrainView.Delete] for further details.
func (v TerrainHandle) Delete() {
	v.view.Delete(v.id)
}

// IsValid returns whether the wrapped terrain has not been deleted.
func (v TerrainHandle) IsValid() bool {
	return v.view.IsValid(v.id)
}

// AttachCollisionMesh attaches a mesh collision shape to the wrapped
// terrain.
//
// See [TerrainView.AttachCollisionMesh] for further details.
func (v TerrainHandle) AttachCollisionMesh(col CollisionMesh) TerrainCollisionShapeID {
	return v.view.AttachCollisionMesh(v.id, col)
}

// DetachCollisionShape removes a collision shape from the wrapped terrain.
//
// See [TerrainView.DetachCollisionShape] for further details.
func (v TerrainHandle) DetachCollisionShape(shapeID TerrainCollisionShapeID) {
	v.view.DetachCollisionShape(v.id, shapeID)
}

type terrainState struct {
	terrainID placement3d.TerrainID
	revision  int32
}

func (s *terrainState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
