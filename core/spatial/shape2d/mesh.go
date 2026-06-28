package shape2d

import "github.com/mokiat/gomath/dprec"

// Mesh represents an arbitrary boundary in 2D space, described as a collection
// of edges.
type Mesh struct {
	// Edges holds the edges that make up the mesh.
	Edges []Edge
}

// NewMesh creates a [Mesh] from the given edges. The slice is retained rather
// than copied, so it should not be modified after the call.
func NewMesh(edges []Edge) Mesh {
	return Mesh{
		Edges: edges,
	}
}

// TransformedMesh returns a new [Mesh] whose edges are the result of applying
// the specified transform to each edge of the given mesh. The original mesh is
// left unmodified.
func TransformedMesh(mesh Mesh, transform Transform) Mesh {
	edges := make([]Edge, len(mesh.Edges))
	for i, edge := range mesh.Edges {
		edges[i] = TransformedEdge(edge, transform)
	}
	return Mesh{
		Edges: edges,
	}
}

// BoundingCircle returns a [Circle] that fully encompasses the mesh.
//
// The circle is centered at the average of all edge endpoints and its radius is
// the distance from that center to the farthest endpoint. The result is
// guaranteed to contain every edge but is not necessarily the smallest possible
// bounding circle; because each edge contributes its own endpoints, the center
// is pulled towards regions that have more edges.
//
// An empty mesh yields the zero [Circle], which is a point of zero radius at the
// origin.
func (m Mesh) BoundingCircle() Circle {
	if len(m.Edges) == 0 {
		return Circle{}
	}

	var center dprec.Vec2
	for _, edge := range m.Edges {
		center = dprec.Vec2Sum(center, edge.A)
		center = dprec.Vec2Sum(center, edge.B)
	}
	center = dprec.Vec2Quot(center, float64(2*len(m.Edges)))

	var radius float64
	for _, edge := range m.Edges {
		radius = max(radius,
			dprec.Sqrt(max(
				dprec.Vec2Diff(edge.A, center).SqrLength(),
				dprec.Vec2Diff(edge.B, center).SqrLength(),
			)),
		)
	}

	return Circle{
		Center: center,
		Radius: radius,
	}
}
