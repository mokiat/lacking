package shape3d

import "github.com/mokiat/gomath/dprec"

// Mesh represents an arbitrary surface in 3D space, described as a collection
// of triangles.
type Mesh struct {
	// Triangles holds the triangles that make up the mesh.
	Triangles []Triangle
}

// NewMesh creates a [Mesh] from the given triangles. The slice is retained
// rather than copied, so it should not be modified after the call.
func NewMesh(triangles []Triangle) Mesh {
	return Mesh{
		Triangles: triangles,
	}
}

// TransformedMesh returns a new [Mesh] whose triangles are the result of applying
// the specified transform to each triangle of the given mesh. The original mesh
// is left unmodified.
func TransformedMesh(mesh Mesh, transform Transform) Mesh {
	result := make([]Triangle, len(mesh.Triangles))
	for i, triangle := range mesh.Triangles {
		result[i] = TransformedTriangle(triangle, transform)
	}
	return Mesh{
		Triangles: result,
	}
}

// BoundingSphere returns a [Sphere] that fully encompasses the mesh.
//
// The sphere is centered at the average of all triangle vertices and its radius
// is the distance from that center to the farthest vertex. The result is
// guaranteed to contain every triangle but is not necessarily the smallest
// possible bounding sphere; because each triangle contributes its own vertices,
// the center is pulled towards regions that are more finely tessellated.
//
// An empty mesh yields the zero [Sphere], which is a point of zero radius at the
// origin.
func (m Mesh) BoundingSphere() Sphere {
	if len(m.Triangles) == 0 {
		return Sphere{}
	}

	var center dprec.Vec3
	for _, triangle := range m.Triangles {
		center = dprec.Vec3Sum(center, triangle.A)
		center = dprec.Vec3Sum(center, triangle.B)
		center = dprec.Vec3Sum(center, triangle.C)
	}
	center = dprec.Vec3Quot(center, float64(3*len(m.Triangles)))

	var radius float64
	for _, triangle := range m.Triangles {
		radius = max(radius,
			dprec.Sqrt(max(
				dprec.Vec3Diff(triangle.A, center).SqrLength(),
				dprec.Vec3Diff(triangle.B, center).SqrLength(),
				dprec.Vec3Diff(triangle.C, center).SqrLength(),
			)),
		)
	}

	return Sphere{
		Center: center,
		Radius: radius,
	}
}
