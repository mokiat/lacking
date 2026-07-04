package isec2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// CheckSegmentRectangle reports whether the directed segment enters the
// rectangle.
//
// Like all segment checks in this package, the test is oriented and
// face-culled: the segment is treated as a directed probe from A to B and only
// a crossing into a front-facing edge (one whose outward normal opposes the
// segment direction) within the segment's extent counts as an intersection. A
// segment that lies entirely inside the rectangle, or that reaches it only
// through a back-facing edge, is not considered to intersect it.
func CheckSegmentRectangle(segment shape2d.Segment, rectangle shape2d.Rectangle) bool {
	delta := dprec.Vec2Diff(segment.B, segment.A)
	relativeStart := dprec.Vec2Diff(segment.A, rectangle.Center)

	rectangleAxisX := rectangle.Rotation.BasisX
	rectangleAxisY := rectangle.Rotation.BasisY

	startX := dprec.Vec2Dot(relativeStart, rectangleAxisX)
	startY := dprec.Vec2Dot(relativeStart, rectangleAxisY)

	deltaX := dprec.Vec2Dot(delta, rectangleAxisX)
	deltaY := dprec.Vec2Dot(delta, rectangleAxisY)

	tCloseX, tFarX, ok := slabRange(startX, deltaX, rectangle.HalfWidth)
	if !ok {
		return false
	}
	tCloseY, tFarY, ok := slabRange(startY, deltaY, rectangle.HalfHeight)
	if !ok {
		return false
	}

	tClose := max(tCloseX, tCloseY)
	tFar := min(tFarX, tFarY)
	return tClose <= tFar && tClose >= 0.0 && tClose <= 1.0
}

// CheckSegmentRectangleOverlap reports whether the segment and the rectangle
// overlap in any way.
//
// Unlike [CheckSegmentRectangle], this test is neither oriented nor face-culled:
// it returns true whenever any part of the segment lies within the rectangle,
// including when the segment lies entirely inside it or starts inside it and
// exits. The result does not depend on the order of the segment's endpoints.
//
// Use it when only the fact of an overlap matters; use [CheckSegmentRectangle]
// or [ResolveSegmentRectangle] when the directed entry point is needed.
func CheckSegmentRectangleOverlap(segment shape2d.Segment, rectangle shape2d.Rectangle) bool {
	delta := dprec.Vec2Diff(segment.B, segment.A)
	relativeStart := dprec.Vec2Diff(segment.A, rectangle.Center)

	rectangleAxisX := rectangle.Rotation.BasisX
	rectangleAxisY := rectangle.Rotation.BasisY

	startX := dprec.Vec2Dot(relativeStart, rectangleAxisX)
	startY := dprec.Vec2Dot(relativeStart, rectangleAxisY)

	deltaX := dprec.Vec2Dot(delta, rectangleAxisX)
	deltaY := dprec.Vec2Dot(delta, rectangleAxisY)

	tCloseX, tFarX, ok := slabRange(startX, deltaX, rectangle.HalfWidth)
	if !ok {
		return false
	}
	tCloseY, tFarY, ok := slabRange(startY, deltaY, rectangle.HalfHeight)
	if !ok {
		return false
	}

	tClose := max(tCloseX, tCloseY)
	tFar := min(tFarX, tFarY)
	return tClose <= tFar && tFar >= 0.0 && tClose <= 1.0
}

// ResolveSegmentRectangle yields the contact at which the directed segment
// enters the rectangle, if it enters it at all.
//
// The contact follows the entry-point convention shared by the segment Resolve
// routines in this package:
//
//   - TargetPoint is the point where the segment first crosses into the
//     rectangle.
//   - TargetNormal is the outward normal of the entered edge.
//   - Depth is the fraction of the segment lying beyond the entry point, in the
//     range [0, 1] (1 when the segment enters at A, 0 when it enters at B). It
//     is comparable across shapes, so [shape2d.DeepestContact] selects the
//     earliest entry along the segment.
func ResolveSegmentRectangle(segment shape2d.Segment, rectangle shape2d.Rectangle, yield shape2d.ContactCallback) {
	delta := dprec.Vec2Diff(segment.B, segment.A)
	relativeStart := dprec.Vec2Diff(segment.A, rectangle.Center)

	rectangleAxisX := rectangle.Rotation.BasisX
	rectangleAxisY := rectangle.Rotation.BasisY

	startX := dprec.Vec2Dot(relativeStart, rectangleAxisX)
	startY := dprec.Vec2Dot(relativeStart, rectangleAxisY)

	deltaX := dprec.Vec2Dot(delta, rectangleAxisX)
	deltaY := dprec.Vec2Dot(delta, rectangleAxisY)

	tCloseX, tFarX, ok := slabRange(startX, deltaX, rectangle.HalfWidth)
	if !ok {
		return
	}
	tCloseY, tFarY, ok := slabRange(startY, deltaY, rectangle.HalfHeight)
	if !ok {
		return
	}
	tClose := max(tCloseX, tCloseY)
	tFar := min(tFarX, tFarY)

	if (tClose > tFar) || (tClose < 0.0) || (tClose > 1.0) {
		return
	}

	contactPoint := dprec.Vec2Lerp(segment.A, segment.B, tClose)

	var normal dprec.Vec2
	switch tClose {
	case tCloseX:
		normal = dprec.Vec2Prod(rectangleAxisX, -dprec.Sign(deltaX))
	case tCloseY:
		normal = dprec.Vec2Prod(rectangleAxisY, -dprec.Sign(deltaY))
	default:
		normal = dprec.BasisXVec2() // should not happen, but just in case
	}

	yield(shape2d.Contact{
		TargetPoint:  contactPoint,
		TargetNormal: normal,
		Depth:        1.0 - tClose,
	})
}
