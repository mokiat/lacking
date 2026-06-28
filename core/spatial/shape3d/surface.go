package shape3d

import "github.com/mokiat/gomath/dprec"

// Surface represents an infinite plane in 3D space, encoded in Hesse normal
// form. The plane is the set of points p for which Dot(Normal, p) equals
// Distance. The Normal is expected to be a unit vector.
type Surface struct {
	// Normal is the unit vector that the plane faces.
	Normal dprec.Vec3
	// Distance is the signed distance of the plane from the origin, measured
	// along the normal.
	Distance float64
}

// BasisXSurface returns a [Surface] that passes through the origin and faces
// along the X axis.
func BasisXSurface() Surface {
	return Surface{
		Normal:   dprec.BasisXVec3(),
		Distance: 0.0,
	}
}

// BasisYSurface returns a [Surface] that passes through the origin and faces
// along the Y axis.
func BasisYSurface() Surface {
	return Surface{
		Normal:   dprec.BasisYVec3(),
		Distance: 0.0,
	}
}

// BasisZSurface returns a [Surface] that passes through the origin and faces
// along the Z axis.
func BasisZSurface() Surface {
	return Surface{
		Normal:   dprec.BasisZVec3(),
		Distance: 0.0,
	}
}

// Point returns the point on the surface that is closest to the origin.
func (s Surface) Point() dprec.Vec3 {
	return dprec.Vec3Prod(s.Normal, s.Distance)
}

// SignedDistance returns the distance of the specified point from the surface,
// measured along the normal. The result is positive when the point lies on the
// side the normal faces, negative on the opposite side, and zero when the point
// lies on the surface.
func (s Surface) SignedDistance(point dprec.Vec3) float64 {
	return dprec.Vec3Dot(s.Normal, point) - s.Distance
}
