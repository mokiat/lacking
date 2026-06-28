package shape2d

import "github.com/mokiat/gomath/dprec"

// Surface represents an infinite line in 2D space, encoded in Hesse normal
// form. The line is the set of points p for which Dot(Normal, p) equals
// Distance, and it splits the plane into a front half (the side the normal
// faces) and a back half.
//
// It is the 2D analogue of the Surface type in the shape3d package. Despite the
// name, kept for symmetry with that package, in two dimensions the shape is a
// line rather than a surface. The Normal is expected to be a unit vector.
type Surface struct {
	// Normal is the unit vector that the line faces.
	Normal dprec.Vec2
	// Distance is the signed distance of the line from the origin, measured
	// along the normal.
	Distance float64
}

// BasisXSurface returns a [Surface] that passes through the origin and faces
// along the X axis.
func BasisXSurface() Surface {
	return Surface{
		Normal:   dprec.BasisXVec2(),
		Distance: 0.0,
	}
}

// BasisYSurface returns a [Surface] that passes through the origin and faces
// along the Y axis.
func BasisYSurface() Surface {
	return Surface{
		Normal:   dprec.BasisYVec2(),
		Distance: 0.0,
	}
}

// Point returns the point on the line that is closest to the origin.
func (s Surface) Point() dprec.Vec2 {
	return dprec.Vec2Prod(s.Normal, s.Distance)
}

// SignedDistance returns the distance of the specified point from the line,
// measured along the normal. The result is positive when the point lies on the
// side the normal faces, negative on the opposite side, and zero when the point
// lies on the line.
func (s Surface) SignedDistance(point dprec.Vec2) float64 {
	return dprec.Vec2Dot(s.Normal, point) - s.Distance
}
