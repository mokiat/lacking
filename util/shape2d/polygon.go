package shape2d

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
)

// NewPolygon creates a polygon from the specified segments.
//
// NOTE: The polygon becomes the owner of the segment slice.
func NewPolygon(segments []Segment) Polygon {
	return Polygon{
		Segments: segments,
	}
}

// TransformedPolygon creates a new polygon from the specified source polygon by
// applying the specified transformation.
func TransformedPolygon(source Polygon, transform Transform) Polygon {
	return Polygon{
		Segments: gog.Map(source.Segments, func(segment Segment) Segment {
			return Segment{
				A: transform.Apply(segment.A),
				B: transform.Apply(segment.B),
			}
		}),
	}
}

// Polygon represents a 2D polygon defined by a series of segments. It is
// more similar to a 3D mesh than a mathematical polygon, as it does not
// enforce any constraints on the segments and they can be disjoint.
type Polygon struct {

	// Segments specifies the segments that make up the polygon.
	Segments []Segment
}

// BoundingCircle returns a Circle that encompases this polygon.
func (p *Polygon) BoundingCircle() Circle {
	if len(p.Segments) == 0 {
		return Circle{}
	}

	var center dprec.Vec2
	for _, segment := range p.Segments {
		center = dprec.Vec2Sum(center, segment.A)
		center = dprec.Vec2Sum(center, segment.B)
	}
	center = dprec.Vec2Quot(center, float64(2*len(p.Segments)))

	var radius float64
	for _, segment := range p.Segments {
		segmentBC := segment.BoundingCircle()
		distance := dprec.Vec2Diff(segmentBC.Position, center)
		radius = max(radius, segmentBC.Radius+distance.Length())
	}

	return NewCircle(center, radius)
}
