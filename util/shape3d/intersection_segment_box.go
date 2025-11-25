package shape3d

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// CheckSegmentBoxIntersection checks if the specified segment intersects
// the specified box and returns the intersection information.
//
// This implementation uses the slab method adapted for oriented boxes.
// It returns the closest intersection point along the segment.
func CheckSegmentBoxIntersection(segment Segment, box Box, yield IntersectionYieldFunc) {
	delta := dprec.Vec3Diff(segment.B, segment.A)
	relativeStart := dprec.Vec3Diff(segment.A, box.Position)

	boxAxisX := box.Rotation.OrientationX()
	boxAxisY := box.Rotation.OrientationY()
	boxAxisZ := box.Rotation.OrientationZ()

	startX := dprec.Vec3Dot(relativeStart, boxAxisX)
	startY := dprec.Vec3Dot(relativeStart, boxAxisY)
	startZ := dprec.Vec3Dot(relativeStart, boxAxisZ)

	dirX := dprec.Vec3Dot(delta, boxAxisX)
	dirY := dprec.Vec3Dot(delta, boxAxisY)
	dirZ := dprec.Vec3Dot(delta, boxAxisZ)

	var (
		tClose = -math.MaxFloat64
		tFar   = math.MaxFloat64
		normal dprec.Vec3
	)

	tLowX := (-box.HalfWidth - startX) / dirX
	tHighX := (box.HalfWidth - startX) / dirX
	tCloseX := min(tLowX, tHighX)
	tFarX := max(tLowX, tHighX)
	if tCloseX > tClose {
		normal = dprec.Vec3Prod(boxAxisX, -dprec.Sign(dirX))
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
		normal = dprec.Vec3Prod(boxAxisY, -dprec.Sign(dirY))
		tClose = tCloseY
	}
	if tFarY < tFar {
		tFar = tFarY
	}

	tLowZ := (-box.HalfLength - startZ) / dirZ
	tHighZ := (box.HalfLength - startZ) / dirZ
	tCloseZ := min(tLowZ, tHighZ)
	tFarZ := max(tLowZ, tHighZ)
	if tCloseZ > tClose {
		normal = dprec.Vec3Prod(boxAxisZ, -dprec.Sign(dirZ))
		tClose = tCloseZ
	}
	if tFarZ < tFar {
		tFar = tFarZ
	}

	if (tClose > tFar) || (tClose < 0.0) || (tClose > 1.0) {
		return
	}

	intersectionPoint := dprec.Vec3Lerp(segment.A, segment.B, tClose)
	depth := dprec.Vec3Dot(
		dprec.Vec3Diff(intersectionPoint, segment.B),
		normal,
	)

	yield(Intersection{
		TargetContact: intersectionPoint,
		TargetNormal:  normal,
		Depth:         depth,
	})
}
