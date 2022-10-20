package shape

import "github.com/mokiat/gomath/dprec"

// NewStaticMesh creates a new StaticMesh from the specified list of
// triangles.
func NewStaticMesh(triangles []StaticTriangle) StaticMesh {
	return StaticMesh{
		Transform: IdentityTransform(),
		triangles: triangles,
		bsRadius:  triangleListBoundingSphereRadius(triangles),
	}
}

// StaticMesh represents an immutable collection of triangles.
type StaticMesh struct {
	Transform
	triangles []StaticTriangle
	bsRadius  float64
}

// BoundingSphereRadius returns the radius of a sphere that can encompass
// this shape.
func (m StaticMesh) BoundingSphereRadius() float64 {
	return m.bsRadius
}

// Triangles returns the list of triangles that make up this mesh. The contents
// of the returned slice must never be modified.
func (m StaticMesh) Triangles() []StaticTriangle {
	return m.triangles
}

// WithTransform returns a new StaticMesh that is based on this one but has
// the specified transform.
func (b StaticMesh) WithTransform(transform Transform) StaticMesh {
	b.Transform = transform
	return b
}

// Transformed returns a new StaticMesh that is based on this one but has
// the specified transform applied to it.
func (b StaticMesh) Transformed(parent Transform) StaticMesh {
	b.Transform = b.Transform.Transformed(parent)
	return b
}

func triangleListBoundingSphereRadius(triangles []StaticTriangle) float64 {
	var radius float64
	for _, triangle := range triangles {
		if pointDistance := dprec.Vec3(triangle.A()).Length(); pointDistance > radius {
			radius = pointDistance
		}
		if pointDistance := dprec.Vec3(triangle.B()).Length(); pointDistance > radius {
			radius = pointDistance
		}
		if pointDistance := dprec.Vec3(triangle.C()).Length(); pointDistance > radius {
			radius = pointDistance
		}
	}
	return radius
}
