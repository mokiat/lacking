package placement3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
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
	Position opt.T[dprec.Vec3]

	// Rotation optionally specifies a rotation of the mesh.
	//
	// Defaults to the identity rotation.
	Rotation opt.T[dprec.Quat]

	// Filtering holds the collision-filtering metadata for the mesh.
	Filtering FilterInfo

	// UserData allows one to attach custom user data to the mesh.
	UserData M

	// Mesh contains the mesh information.
	//
	// The mesh must have at least one triangle. An empty mesh has no area to
	// be placed in the scene and is considered invalid.
	Mesh shape3d.Mesh
}

type meshShape[M any] struct {
	spatialID query3d.TreeItemID
	filterRepresentation
	meshRepresentation
	userData M
}

func shapeMeshCanIntersect[S, M any](shape *shape[S], mesh *meshShape[M]) bool {
	return shape.canInteractWith(&mesh.filterRepresentation)
}

// TODO: Consider using a different storage mechanism. For example an
// Octree or BVH structure.
// Alternatively experiment with placing each mesh triangle in the existing
// mesh tree, through this will likely destroy the mesh tree performance.

type meshRepresentation struct {
	wsBSphere   shape3d.Sphere
	wsAABB      shape3d.AABB
	wsTriangles []shape3d.Triangle
}

func newMeshRepresentation(mesh shape3d.Mesh) meshRepresentation {
	return meshRepresentation{
		wsBSphere:   mesh.BoundingSphere(),
		wsAABB:      mesh.BoundingAABB(),
		wsTriangles: mesh.Triangles,
	}
}
