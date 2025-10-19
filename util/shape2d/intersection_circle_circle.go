package shape2d

import "github.com/mokiat/gomath/dprec"

// IsCircleCircleIntersection checks if the source circle intersects with
// the target circle.
//
// Only a bool result is returned and no collision points or separation
// normals are evaluated.
func IsCircleCircleIntersection(source, target Circle) bool {
	distance := dprec.Vec2Diff(source.Position, target.Position).Length()
	return distance <= (source.Radius + target.Radius)
}

// CheckCircleCircleIntersection checks if a Circle shape intersects with
// another Circle shape.
func CheckCircleCircleIntersection(source, target Circle) (Intersection, bool) {
	sourcePosition := source.Position
	sourceRadius := source.Radius

	targetPosition := target.Position
	targetRadius := target.Radius

	deltaPosition := dprec.Vec2Diff(targetPosition, sourcePosition)
	distance := deltaPosition.Length()

	overlap := (sourceRadius + targetRadius) - distance
	if overlap <= 0.0 {
		return Intersection{}, false
	}

	targetNormal := dprec.Vec2Quot(deltaPosition, -distance) // unit vector
	return Intersection{
		TargetContact: dprec.Vec2Sum(
			targetPosition,
			dprec.Vec2Prod(targetNormal, targetRadius),
		),
		TargetNormal: targetNormal,
		Depth:        overlap,
	}, true
}
