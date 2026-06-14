package shape2d

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
)

// NewPolygon creates a polygon from the specified edges.
//
// NOTE: The polygon becomes the owner of the segment slice.
func NewPolygon(edges []Edge) Polygon {
	return Polygon{
		Edges: edges,
	}
}

// TransformedPolygon creates a new polygon from the specified source polygon by
// applying the specified transformation.
func TransformedPolygon(source Polygon, transform Transform) Polygon {
	return Polygon{
		Edges: gog.Map(source.Edges, func(segment Edge) Edge {
			return Edge{
				A: transform.Apply(segment.A),
				B: transform.Apply(segment.B),
			}
		}),
	}
}

// Polygon represents a 2D polygon defined by a series of edges. It is
// more similar to a 3D mesh than a mathematical polygon as it does not
// enforce any constraints on the edges and they can be disjoint.
type Polygon struct {

	// Edges specifies the segments that make up the polygon.
	Edges []Edge
}

// BoundingCircle returns a Circle that encompases this polygon.
func (p Polygon) BoundingCircle() Circle {
	if len(p.Edges) == 0 {
		return Circle{}
	}

	var center dprec.Vec2
	for _, segment := range p.Edges {
		center = dprec.Vec2Sum(center, segment.A)
		center = dprec.Vec2Sum(center, segment.B)
	}
	center = dprec.Vec2Quot(center, float64(2*len(p.Edges)))

	var radius float64
	for _, segment := range p.Edges {
		segmentBC := segment.BoundingCircle()
		distance := dprec.Vec2Diff(segmentBC.Position, center)
		radius = max(radius, segmentBC.Radius+distance.Length())
	}

	return NewCircle(center, radius)
}
