package shape

func NewStaticMesh(triangles []StaticTriangle) StaticMesh {
	return StaticMesh{
		triangles: triangles,
		radius:    triangleListBoundingSphereRadius(triangles),
	}
}

type StaticMesh struct {
	triangles []StaticTriangle
	radius    float32
}

func (m StaticMesh) Triangles() []StaticTriangle {
	return m.triangles
}

func (m StaticMesh) BoundingSphereRadius() float32 {
	return m.radius
}

func triangleListBoundingSphereRadius(triangles []StaticTriangle) float32 {
	var radius float32
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
