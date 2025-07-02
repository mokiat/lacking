package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// CheckIntersectionSphereWithSphere checks if a Sphere shape intersects with
// another Sphere shape.
func CheckIntersectionSphereWithSphere(first, second Sphere) opt.T[Intersection] {
	firstPosition := first.Position
	firstRadius := first.Radius

	secondPosition := second.Position
	secondRadius := second.Radius

	deltaPosition := dprec.Vec3Diff(secondPosition, firstPosition)
	distance := deltaPosition.Length()

	overlap := (firstRadius + secondRadius) - distance
	if overlap <= 0.0 {
		return opt.Unspecified[Intersection]()
	}

	secondDisplaceNormal := dprec.Vec3Quot(deltaPosition, distance) // unit vec
	firstDisplaceNormal := dprec.InverseVec3(secondDisplaceNormal)

	return opt.V(Intersection{
		Depth: overlap,
		FirstContact: dprec.Vec3Sum(
			firstPosition,
			dprec.Vec3Prod(secondDisplaceNormal, firstRadius),
		),
		FirstDisplaceNormal: firstDisplaceNormal,
		SecondContact: dprec.Vec3Sum(
			secondPosition,
			dprec.Vec3Prod(firstDisplaceNormal, secondRadius),
		),
		SecondDisplaceNormal: secondDisplaceNormal,
	})
}
