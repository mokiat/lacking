package placement3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d"
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

	// ShapeInfo contains general shape information.
	//
	// TODO: Rename ShapeInfo to something more meaningful for both
	// shapes and meshes.
	ShapeInfo[M]

	// Mesh contains the mesh information.
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

type meshRepresentation struct {
	wsBSphere   shape3d.Sphere
	wsTriangles []shape3d.Triangle

	// TODO: Move this into scene!
	points [3]dprec.Vec3
}

func newMeshRepresentation(mesh shape3d.Mesh) meshRepresentation {
	return meshRepresentation{
		wsBSphere:   mesh.BoundingSphere(),
		wsTriangles: mesh.Triangles,
		points:      [3]dprec.Vec3{}, // just used for GJK to avoid allocations
	}
}

func (s *meshRepresentation) boundingSphere() shape3d.Sphere {
	return s.wsBSphere
}

func (s *meshRepresentation) gjkShapeCount() int {
	return len(s.wsTriangles)
}

func (s *meshRepresentation) gjkShape(index int) gjk3d.Shape {
	triangle := &s.wsTriangles[index]
	points := s.points[:]
	points[0] = triangle.A
	points[1] = triangle.B
	points[2] = triangle.C
	return gjk3d.Shape{
		Position:   dprec.ZeroVec3(),
		Rotation:   shape3d.IdentityRotation(),
		Points:     points,
		SkinRadius: 0.0,
	}
}
