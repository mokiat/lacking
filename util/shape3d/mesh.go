package shape3d

import "github.com/mokiat/gomath/dprec"

func NewMesh(triangles []Triangle) Mesh {
	return Mesh{
		Triangles: triangles,
	}
}

type Mesh struct {
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
