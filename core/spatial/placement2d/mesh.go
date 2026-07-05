package placement2d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// InvalidMeshID indicates a mesh that can never be part of the scene.
const InvalidMeshID = MeshID(nilIndex)

// MeshID is a reference to a mesh in the scene.
type MeshID int32

// MeshInfo contains the information needed to create a mesh shape.
type MeshInfo[M any] struct {

	// Position optionally specifies a position where the mesh should be placed.
	//
	// Defaults to the origin.
	Position opt.T[dprec.Vec2]

	// Rotation optionally specifies a rotation of the mesh.
	//
	// Defaults to the identity rotation.
	Rotation opt.T[dprec.Angle]

	// Filtering holds the collision-filtering metadata for the mesh.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the mesh.
	UserData M

	// Mesh contains the mesh information.
	Mesh shape2d.Mesh
}

type meshShape[M any] struct {
	spatialID query2d.TreeItemID
	filterRepresentation
	meshRepresentation
	userData M
}

func shapeMeshCanIntersect[S, M any](shape *shape[S], mesh *meshShape[M]) bool {
	return shape.canInteractWith(&mesh.filterRepresentation)
}

type meshRepresentation struct {
	wsBCircle shape2d.Circle

	// TODO: Consider using a different storage mechanism. For example a
	// Quadtree or BVH structure.
	// Alternatively experiment with placing each mesh edge in the existing
	// mesh tree, through this will likely destroy the mesh tree performance.
	wsEdges []shape2d.Edge
}

func newMeshRepresentation(mesh shape2d.Mesh) meshRepresentation {
	return meshRepresentation{
		wsBCircle: mesh.BoundingCircle(),
		wsEdges:   mesh.Edges,
	}
}

func (s *meshRepresentation) boundingCircle() shape2d.Circle {
	return s.wsBCircle
}
