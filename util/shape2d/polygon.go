package shape2d

func NewPolygon(segments []Segment) Polygon {
	return Polygon{
		Segments: segments,
	}
}

type Polygon struct {
	Segments []Segment
}
