package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSegmentBox reports whether the directed segment enters the box.
//
// Like all segment checks in this package, the test is oriented and
// face-culled: the segment is treated as a directed probe from A to B and only
// a crossing into a front-facing face (one whose outward normal opposes the
// segment direction) within the segment's extent counts as an intersection. A
// segment that lies entirely inside the box, or that reaches it only through a
// back-facing face, is not considered to intersect it.
func CheckSegmentBox(segment shape3d.Segment, box shape3d.Box) bool {
	delta := dprec.Vec3Diff(segment.B, segment.A)
	relativeStart := dprec.Vec3Diff(segment.A, box.Center)

	boxAxisX := box.Rotation.BasisX
	boxAxisY := box.Rotation.BasisY
	boxAxisZ := box.Rotation.BasisZ

	startX := dprec.Vec3Dot(relativeStart, boxAxisX)
	startY := dprec.Vec3Dot(relativeStart, boxAxisY)
	startZ := dprec.Vec3Dot(relativeStart, boxAxisZ)

	deltaX := dprec.Vec3Dot(delta, boxAxisX)
	deltaY := dprec.Vec3Dot(delta, boxAxisY)
	deltaZ := dprec.Vec3Dot(delta, boxAxisZ)

	tCloseX, tFarX, ok := slabRange(startX, deltaX, box.HalfWidth)
	if !ok {
		return false
	}
	tCloseY, tFarY, ok := slabRange(startY, deltaY, box.HalfHeight)
	if !ok {
		return false
	}
	tCloseZ, tFarZ, ok := slabRange(startZ, deltaZ, box.HalfLength)
	if !ok {
		return false
	}

	tClose := max(tCloseX, tCloseY, tCloseZ)
	tFar := min(tFarX, tFarY, tFarZ)
	return tClose <= tFar && tClose >= 0.0 && tClose <= 1.0
}

// ResolveSegmentBox yields the contact at which the directed segment enters the
// box, if it enters one at all.
//
// The contact follows the entry-point convention shared by the segment Resolve
// routines in this package:
//
//   - TargetPoint is the point where the segment first crosses into the box.
//   - TargetNormal is the outward normal of the entered face.
//   - Depth is how far the far endpoint B has travelled past that face along
//     the normal, so moving the segment by Depth along TargetNormal brings B
//     back onto the entry face.
func ResolveSegmentBox(segment shape3d.Segment, box shape3d.Box, yield shape3d.ContactCallback) {
	delta := dprec.Vec3Diff(segment.B, segment.A)
	relativeStart := dprec.Vec3Diff(segment.A, box.Center)

	boxAxisX := box.Rotation.BasisX
	boxAxisY := box.Rotation.BasisY
	boxAxisZ := box.Rotation.BasisZ

	startX := dprec.Vec3Dot(relativeStart, boxAxisX)
	startY := dprec.Vec3Dot(relativeStart, boxAxisY)
	startZ := dprec.Vec3Dot(relativeStart, boxAxisZ)

	deltaX := dprec.Vec3Dot(delta, boxAxisX)
	deltaY := dprec.Vec3Dot(delta, boxAxisY)
	deltaZ := dprec.Vec3Dot(delta, boxAxisZ)

	tCloseX, tFarX, ok := slabRange(startX, deltaX, box.HalfWidth)
	if !ok {
		return
	}
	tCloseY, tFarY, ok := slabRange(startY, deltaY, box.HalfHeight)
	if !ok {
		return
	}
	tCloseZ, tFarZ, ok := slabRange(startZ, deltaZ, box.HalfLength)
	if !ok {
		return
	}
	tClose := max(tCloseX, tCloseY, tCloseZ)
	tFar := min(tFarX, tFarY, tFarZ)

	if (tClose > tFar) || (tClose < 0.0) || (tClose > 1.0) {
		return
	}

	var normal dprec.Vec3
	switch tClose {
	case tCloseX:
		normal = dprec.Vec3Prod(boxAxisX, -dprec.Sign(deltaX))
	case tCloseY:
		normal = dprec.Vec3Prod(boxAxisY, -dprec.Sign(deltaY))
	case tCloseZ:
		normal = dprec.Vec3Prod(boxAxisZ, -dprec.Sign(deltaZ))
	default:
		normal = dprec.BasisXVec3() // should not happen, but just in case
	}

	contactPoint := dprec.Vec3Lerp(segment.A, segment.B, tClose)
	depth := dprec.Vec3Dot(dprec.Vec3Diff(contactPoint, segment.B), normal)

	yield(shape3d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: normal,
		Depth:        depth,
	})
}
