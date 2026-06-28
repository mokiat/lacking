package isec2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// CheckCircleCircle reports whether the two circles intersect.
//
// Circles that merely touch, or where one fully contains the other, are
// considered to intersect.
func CheckCircleCircle(first, second shape2d.Circle) bool {
	// Compare squared distances to avoid the square root.
	radiusSum := first.Radius + second.Radius
	delta := dprec.Vec2Diff(first.Center, second.Center)
	return delta.SqrLength() <= radiusSum*radiusSum
}

// ResolveCircleCircle yields a [shape2d.Contact] for the overlap of the two
// circles, if there is one.
//
// The contact is reported with the first circle as the source and the second as
// the target: TargetPoint is the point on the second circle's perimeter closest
// to the first, TargetNormal is the outward normal there (pointing toward the
// first circle), and Depth is the overlap of the two circles along that normal.
func ResolveCircleCircle(first, second shape2d.Circle, yield shape2d.ContactCallback) {
	delta := dprec.Vec2Diff(first.Center, second.Center)
	distance := delta.Length()

	overlap := (first.Radius + second.Radius) - distance
	if overlap < 0.0 {
		return
	}

	var normal dprec.Vec2
	if distance == 0 {
		normal = dprec.BasisXVec2()
	} else {
		normal = dprec.Vec2Quot(delta, distance)
	}

	yield(shape2d.Contact{
		TargetPoint:  dprec.Vec2Sum(second.Center, dprec.Vec2Prod(normal, second.Radius)),
		TargetNormal: normal,
		Depth:        overlap,
	})
}
