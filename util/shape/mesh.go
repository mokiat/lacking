package shape

// NewStaticMesh creates a new StaticMesh from the specified list of
// triangles.
func NewStaticMesh(triangles []StaticTriangle) StaticMesh {
	return StaticMesh{
		triangles: triangles,
		radius:    triangleListBoundingSphereRadius(triangles),
	}
}

// StaticMesh represents a collection of triangles that cannot be
// resized or repositioned.
type StaticMesh struct {
	triangles []StaticTriangle
	radius    float64
}

// BoundingSphereRadius returns the radius of a sphere that can encompass
// this shape.
func (m StaticMesh) BoundingSphereRadius() float64 {
	return m.radius
}

// Triangles returns the list of triangles that make up this mesh. The contents
// of the returned slice must never be modified.
func (m StaticMesh) Triangles() []StaticTriangle {
	return m.triangles
}

func triangleListBoundingSphereRadius(triangles []StaticTriangle) float64 {
	var radius float64
	for _, triangle := range triangles {
		if pointDistance := triangle.A().Length(); pointDistance > radius {
			radius = pointDistance
		}
		if pointDistance := triangle.B().Length(); pointDistance > radius {
			radius = pointDistance
		}
		if pointDistance := triangle.C().Length(); pointDistance > radius {
			radius = pointDistance
		}
	}
	return radius
}
