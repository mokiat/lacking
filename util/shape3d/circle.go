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

	// Position specifies the position of the circle.
	Position dprec.Vec3

	// Normal specifies the orientation of the circle using a normal vector.
	Normal dprec.Vec3

	// Radius specifies the radius of the circle.
	Radius float64
}
