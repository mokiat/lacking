package collision

import "github.com/mokiat/gomath/dprec"

// NewMesh creates a new Mesh from the specified list of triangles.
//
// NOTE: The Mesh becomes the owner of the slice so callers should not
// keep a reference to it afterwards or modify the contents in any way.
func NewMesh(triangles []Triangle) Mesh {
	var center dprec.Vec3
	for _, triangle := range triangles {
		center = dprec.Vec3Sum(center, triangle.a)
		center = dprec.Vec3Sum(center, triangle.b)
		center = dprec.Vec3Sum(center, triangle.c)
	}
	center = dprec.Vec3Quot(center, float64(3*len(triangles)))

	var radius float64
	for _, triangle := range triangles {
		triangleBS := triangle.BoundingSphere()
		distance := dprec.Vec3Diff(triangleBS.position, center)
		radius = dprec.Max(radius, triangleBS.radius+distance.Length())
	}

	return Mesh{
		triangles: triangles,
		bs:        NewSphere(center, radius),
	}
}

// Mesh represents a collection of triangles.
type Mesh struct {
	triangles []Triangle
	bs        Sphere
}

// Replace replaces this shape with the template one after the specified
// transformation has been applied to it.
func (m *Mesh) Replace(template Mesh, transform Transform) {
	if len(m.triangles) != len(template.triangles) {
		m.triangles = make([]Triangle, len(template.triangles))
	}
	for i := range m.triangles {
		m.triangles[i].Replace(template.triangles[i], transform)
	}
	m.bs.Replace(template.bs, transform)
}

// Triangles returns the list of triangles that make up this mesh.
//
// NOTE: This returns the internal slice of triangles so callers should
// not modify the contents in any way.
func (m *Mesh) Triangles() []Triangle {
	return m.triangles
}

// BoundingSphere returns a Sphere that encompases this mesh.
func (m *Mesh) BoundingSphere() Sphere {
	return m.bs
}
