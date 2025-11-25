package shape3d

import "github.com/mokiat/gomath/dprec"

// NewSurface returns a new surface from the specified point and normal.
func NewSurface(point, normal dprec.Vec3) Surface {
	return Surface{
		Normal:   normal,
		Distance: dprec.Vec3Dot(point, normal),
	}
}

// BasisXSurface surface returns a surface that is oriented along the X axis.
func BasisXSurface() Surface {
	return Surface{
		Normal:   dprec.BasisXVec3(),
		Distance: 0.0,
	}
}

// BasisYSurface surface returns a surface that is oriented along the Y axis.
func BasisYSurface() Surface {
	return Surface{
		Normal:   dprec.BasisYVec3(),
		Distance: 0.0,
	}
}

// BasisZSurface surface returns a surface that is oriented along the Z axis.
func BasisZSurface() Surface {
	return Surface{
		Normal:   dprec.BasisZVec3(),
		Distance: 0.0,
	}
}

// Surface represents a plane in 3D space.
//
// It uses a normal + distance representation which both compact and also
// easier to use that the official A, B, C, D, since the normal is expected
// to be normalized.
type Surface struct {

	// Normal is a unit vector that represents the orientation of the plane.
	Normal dprec.Vec3

	// Distance is the distance of the plane from the origin along the normal.
	Distance float64
}

// Point returns an arbitrary point on the surface.
func (s Surface) Point() dprec.Vec3 {
	return dprec.Vec3Prod(s.Normal, s.Distance)
}
