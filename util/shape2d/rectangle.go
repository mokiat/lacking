package shape2d

import "github.com/mokiat/gomath/dprec"

func NewRectangle(pos dprec.Vec2, rotation dprec.Angle, width, height float64) Rectangle {
	return Rectangle{
		Position:   pos,
		Rotation:   rotation,
		HalfWidth:  width / 2.0,
		HalfHeight: height / 2.0,
	}
}

type Rectangle struct {
	Position   dprec.Vec2
	Rotation   dprec.Angle
	HalfWidth  float64
	HalfHeight float64
}
