package shape3d

import "github.com/mokiat/gomath/dprec"

func IsSegmentAABBIntersection(segment Segment, aabb AABB) bool {
	dir := dprec.Vec3Diff(segment.B, segment.A)

	tLowX := (aabb.MinX - segment.A.X) / dir.X
	tHighX := (aabb.MaxX - segment.A.X) / dir.X
	tCloseX := min(tLowX, tHighX)
	tFarX := max(tLowX, tHighX)

	tLowY := (aabb.MinY - segment.A.Y) / dir.Y
	tHighY := (aabb.MaxY - segment.A.Y) / dir.Y
	tCloseY := min(tLowY, tHighY)
	tFarY := max(tLowY, tHighY)

	tLowZ := (aabb.MinZ - segment.A.Z) / dir.Z
	tHighZ := (aabb.MaxZ - segment.A.Z) / dir.Z
	tCloseZ := min(tLowZ, tHighZ)
	tFarZ := max(tLowZ, tHighZ)

	tClose := max(tCloseX, tCloseY, tCloseZ)
	tFar := min(tFarX, tFarY, tFarZ)

	return tClose <= tFar && (tClose >= 0 && tClose <= 1.0)
}
