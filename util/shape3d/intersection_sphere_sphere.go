package shape3d

import (
	"github.com/mokiat/gomath/dprec"
)

// IsSphereSphereIntersection checks if the source sphere intersects with
// the target sphere.
//
// Only a bool result is returned and no collision points or separation
// normals are evaluated.
func IsSphereSphereIntersection(source, target Sphere) bool {
	distance := dprec.Vec3Diff(source.Position, target.Position).Length()
	return distance <= (source.Radius + target.Radius)
}

// CheckSphereSphereIntersection checks if a Sphere shape intersects with
// another Sphere shape.
func CheckSphereSphereIntersection(source, target Sphere) (Intersection, bool) {
	sourcePosition := source.Position
	sourceRadius := source.Radius

	targetPosition := target.Position
	targetRadius := target.Radius

	deltaPosition := dprec.Vec3Diff(targetPosition, sourcePosition)
	distance := deltaPosition.Length()

	overlap := (sourceRadius + targetRadius) - distance
	if overlap <= 0.0 {
		return Intersection{}, false
	}

	targetNormal := dprec.Vec3Quot(deltaPosition, -distance) // unit vector
	return Intersection{
		TargetContact: dprec.Vec3Sum(
			targetPosition,
			dprec.Vec3Prod(targetNormal, targetRadius),
		),
		TargetNormal: targetNormal,
		Depth:        overlap,
	}, true
}
