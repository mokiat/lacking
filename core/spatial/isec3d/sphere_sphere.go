package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSphereSphere reports whether the two spheres intersect.
//
// Spheres that merely touch, or where one fully contains the other, are
// considered to intersect.
func CheckSphereSphere(first, second shape3d.Sphere) bool {
	// Compare squared distances to avoid the square root.
	radiusSum := first.Radius + second.Radius
	return dprec.Vec3Diff(first.Center, second.Center).SqrLength() <= dprec.Sqr(radiusSum)
}

// ResolveSphereSphere yields a Contact for the overlap of the two spheres, if
// there is one.
//
// The contact is reported with the first sphere as the source and the second as
// the target: TargetPoint is the point on the second sphere's surface closest to
// the first, TargetNormal is the outward surface normal there (pointing toward
// the first sphere), and Depth is the overlap of the two spheres along that
// normal.
func ResolveSphereSphere(first, second shape3d.Sphere, yield shape3d.ContactCallback) {
	delta := dprec.Vec3Diff(first.Center, second.Center)
	distance := delta.Length()

	radiusSum := first.Radius + second.Radius
	if distance > radiusSum {
		return // the spheres do not reach each other
	}

	var normal dprec.Vec3
	if distance > 0 {
		normal = dprec.Vec3Quot(delta, distance)
	} else {
		// The centers coincide; the separation normal is not unique, so pick a
		// deterministic one.
		normal = dprec.BasisXVec3()
	}

	yield(shape3d.Contact{
		TargetPoint:  dprec.Vec3Sum(second.Center, dprec.Vec3Prod(normal, second.Radius)),
		TargetNormal: normal,
		Depth:        radiusSum - distance,
	})
}
