package shape2d

import "github.com/mokiat/gomath/dprec"

func NewSegment(a, b dprec.Vec2) Segment {
	return Segment{
		A: a,
		B: b,
	}
}

type Segment struct {
	A dprec.Vec2
	B dprec.Vec2
}
