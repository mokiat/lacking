package shape2d

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// CheckSegmentRectangleIntersection checks if the specified segment intersects
// the specified rectangle and returns the intersection information.
//
// This implementation uses the slab method adapted for oriented rectangles.
// It returns the closest intersection point along the segment.
func CheckSegmentRectangleIntersection(segment Segment, box Rectangle, yield IntersectionYieldFunc) {
	delta := dprec.Vec2Diff(segment.B, segment.A)
	relativeStart := dprec.Vec2Diff(segment.A, box.Position)

	cs := dprec.Cos(box.Rotation)
	sn := dprec.Sin(box.Rotation)
	boxAxisX := dprec.NewVec2(cs, sn)
	boxAxisY := dprec.NewVec2(-sn, cs)

	startX := dprec.Vec2Dot(relativeStart, boxAxisX)
	startY := dprec.Vec2Dot(relativeStart, boxAxisY)

	dirX := dprec.Vec2Dot(delta, boxAxisX)
	dirY := dprec.Vec2Dot(delta, boxAxisY)

	var (
		tClose = -math.MaxFloat64
		tFar   = math.MaxFloat64
		normal dprec.Vec2
	)

	tLowX := (-box.HalfWidth - startX) / dirX
	tHighX := (box.HalfWidth - startX) / dirX
	tCloseX := min(tLowX, tHighX)
	tFarX := max(tLowX, tHighX)
	if tCloseX > tClose {
		normal = dprec.Vec2Prod(boxAxisX, -dprec.Sign(dirX))
		tClose = tCloseX
	}
	if tFarX < tFar {
		tFar = tFarX
	}

	tLowY := (-box.HalfHeight - startY) / dirY
	tHighY := (box.HalfHeight - startY) / dirY
	tCloseY := min(tLowY, tHighY)
	tFarY := max(tLowY, tHighY)
	if tCloseY > tClose {
		normal = dprec.Vec2Prod(boxAxisY, -dprec.Sign(dirY))
		tClose = tCloseY
	}
	if tFarY < tFar {
		tFar = tFarY
	}

	if (tClose > tFar) || (tClose < 0.0) || (tClose > 1.0) {
		return
	}

	intersectionPoint := dprec.Vec2Lerp(segment.A, segment.B, tClose)
	depth := dprec.Vec2Dot(
		dprec.Vec2Diff(intersectionPoint, segment.B),
		normal,
	)

	yield(Intersection{
		TargetContact: intersectionPoint,
		TargetNormal:  normal,
		Depth:         depth,
	})
}
