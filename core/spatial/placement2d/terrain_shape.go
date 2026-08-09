package placement2d

import (
	"github.com/mokiat/lacking/core/spatial/query2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// NilTerrainShapeID indicates a terrain shape that can never be part of the
// scene.
const NilTerrainShapeID = TerrainShapeID(nilIndex)

// TerrainShapeID is a reference to a concave shape that is attached to a
// terrain in the scene.
type TerrainShapeID int32

// MeshInfo contains the information needed to create a mesh shape.
type MeshInfo[S any] struct {

	// Filtering holds the collision-filtering metadata for the mesh.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the mesh.
	UserData S

	// Mesh contains the mesh information.
	//
	// The edges of the mesh are specified in world space, since terrains have
	// no transform of their own. Use [shape2d.TransformedMesh] to place a mesh
	// that is modeled around the origin.
	//
	// The edge slice is retained rather than copied, so it must not be
	// modified afterwards.
	//
	// The mesh must have at least one edge. An empty mesh has no area to be
	// placed in the scene and attaching it panics.
	Mesh shape2d.Mesh
}

type terrainShapeState[S any] struct {
	terrainIndex   int32
	nextShapeIndex int32
	prevShapeIndex int32
	spatialID      query2d.TreeItemID
	filterRepresentation
	terrainShapeRepresentation
	userData S
}

// objectTerrainShapesCanIntersect reports whether the specified object shape
// and terrain shape are allowed to be checked for intersection.
func objectTerrainShapesCanIntersect[S any](objectShape *objectShapeState[S], terrainShape *terrainShapeState[S]) bool {
	return objectShape.canInteractWith(&terrainShape.filterRepresentation)
}

// TODO: Consider using a different storage mechanism. For example a
// Quadtree or BVH structure.
//
// TODO: Consider abstracting the edges through a resolver that can find
// candidate edges for bounding circles, allowing for heightline or other
// implementations (types of shapes).

// terrainShapeRepresentation holds the world-space geometry of a terrain
// shape. As terrains cannot be relocated, there is no local-space counterpart
// and the representation never needs to be updated after construction.
type terrainShapeRepresentation struct {
	wsBCircle shape2d.Circle
	wsAABB    shape2d.AABB
	wsEdges   []shape2d.Edge
}
