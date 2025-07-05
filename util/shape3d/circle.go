package shape3d

import "github.com/mokiat/gomath/dprec"

// NewCircle returns a circle from the specified position, normal and radius.
func NewCircle(position, normal dprec.Vec3, radius float64) Circle {
	return Circle{
		Position: position,
		Normal:   normal,
		Radius:   radius,
	}
}

// Circle represents a 2D circle in 3D space.
type Circle struct {
	Position dprec.Vec3
	Normal   dprec.Vec3
	Radius   float64
}
