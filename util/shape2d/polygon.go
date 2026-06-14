package shape2d

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
)

// Polygon represents a 2D polygon defined by a series of edges. It is
// more similar to a 3D mesh than a mathematical polygon as it does not
// enforce any constraints on the edges and they can be disjoint.
type Polygon struct {

	// Edges specifies the segments that make up the polygon.
	Edges []Edge
}

// NewPolygon creates a polygon from the specified edges.
//
// NOTE: The polygon becomes the owner of the edge slice.
func NewPolygon(edges []Edge) Polygon {
	return Polygon{
		Edges: edges,
	}
}

// TransformedPolygon creates a new polygon from the specified source polygon by
// applying the specified transformation.
func TransformedPolygon(source Polygon, transform Transform) Polygon {
	basisTransform := transform.Basis()
	return BasisTransformedPolygon(source, basisTransform)
}

// BasisTransformedPolygon creates a new polygon from the specified source polygon by
// applying the specified basis transformation.
func BasisTransformedPolygon(source Polygon, basisTransform BasisTransform) Polygon {
	return Polygon{
		Edges: gog.Map(source.Edges, func(edge Edge) Edge {
			return BasisTransformedEdge(edge, basisTransform)
		}),
	}
}

// BoundingCircle returns a Circle that encompasses this polygon.
func (p Polygon) BoundingCircle() Circle {
	if len(p.Edges) == 0 {
		return Circle{}
	}

	var center dprec.Vec2
	for _, edge := range p.Edges {
		center = dprec.Vec2Sum(center, edge.A)
		center = dprec.Vec2Sum(center, edge.B)
	}
	center = dprec.Vec2Quot(center, float64(2*len(p.Edges)))

	var radius float64
	for _, edge := range p.Edges {
		edgeBC := edge.BoundingCircle()
		distance := dprec.Vec2Diff(edgeBC.Position, center)
		radius = max(radius, edgeBC.Radius+distance.Length())
	}

	return NewCircle(center, radius)
}
