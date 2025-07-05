package shape3d

import "github.com/mokiat/gomath/dprec"

func NewSegment(a, b dprec.Vec3) Segment {
	return Segment{
		A: a,
		B: b,
	}
}

type Segment struct {
	A dprec.Vec3
	B dprec.Vec3
}
