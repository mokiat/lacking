package shape3d

import (
	"github.com/mokiat/gomath/dprec"
)

func IsSegmentAABBIntersection(segment Segment, aabb AABB, inner bool) bool {
	delta := dprec.Vec3Diff(segment.B, segment.A)
	length := delta.Length()
	dir := dprec.Vec3Quot(delta, length)

	tLowX := (aabb.MinX - segment.A.X) / dir.X
	tLowY := (aabb.MinY - segment.A.Y) / dir.Y
	tLowZ := (aabb.MinZ - segment.A.Z) / dir.Z

	tHighX := (aabb.MaxX - segment.A.X) / dir.X
	tHighY := (aabb.MaxY - segment.A.Y) / dir.Y
	tHighZ := (aabb.MaxZ - segment.A.Z) / dir.Z

	tCloseX := min(tLowX, tHighX)
	tCloseY := min(tLowY, tHighY)
	tCloseZ := min(tLowZ, tHighZ)
	tClose := max(tCloseX, tCloseY, tCloseZ)

	tFarX := max(tLowX, tHighX)
	tFarY := max(tLowY, tHighY)
	tFarZ := max(tLowZ, tHighZ)
	tFar := min(tFarX, tFarY, tFarZ)

	return tClose <= tFar &&
		((tClose >= 0 && tClose <= length) || (inner && tFar >= 0 && tFar <= length))
}
