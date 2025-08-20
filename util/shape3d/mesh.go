package shape3d

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
)

// NewMesh creates a mesh from the specified triangles.
//
// NOTE: The mesh becomes the owner of the triangle slice.
func NewMesh(triangles []Triangle) Mesh {
	return Mesh{
		Triangles: triangles,
	}
}

// TransformedMesh creates a new mesh from the specified source mesh by
// applying the specified transformation.
func TransformedMesh(source Mesh, transform Transform) Mesh {
	return Mesh{
		Triangles: gog.Map(source.Triangles, func(triangle Triangle) Triangle {
			return Triangle{
				A: transform.Apply(triangle.A),
				B: transform.Apply(triangle.B),
				C: transform.Apply(triangle.C),
			}
		}),
	}
}

// Mesh represents a mesh shape that is comprised of triangles.
type Mesh struct {

	// Triangles contains all the triangles that make up the mesh.
	Triangles []Triangle
}

// BoundingSphere returns a Sphere that encompases this mesh.
func (m *Mesh) BoundingSphere() Sphere {
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
		triangleBS := triangle.BoundingSphere()
		distance := dprec.Vec3Diff(triangleBS.Position, center)
		radius = dprec.Max(radius, triangleBS.Radius+distance.Length())
	}

	return NewSphere(center, radius)
}
