package isec3d

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// boxTriangleAxisEpsilon is the squared-length threshold below which a candidate
// separating axis is treated as degenerate and skipped. Such axes arise when a
// box edge is parallel to a triangle edge, making their cross product vanish.
const boxTriangleAxisEpsilon = 1e-15

// CheckBoxTriangle reports whether the box intersects the triangle.
//
// The triangle is treated as a one-sided, front-face-culled surface, as in
// [CheckSphereTriangle]: only a box whose center lies strictly in front of the
// triangle's plane (on the side its normal faces) can intersect it. A box
// centered behind the triangle, or exactly on its plane, is never considered
// to intersect it, even if it would otherwise overlap.
//
// The test itself is the Separating Axis Theorem applied to the box's three
// face normals, the triangle's normal, and the nine
// box-edge-cross-triangle-edge axes. A box that merely touches the triangle
// counts as intersecting.
func CheckBoxTriangle(box shape3d.Box, triangle shape3d.Triangle) bool {
	_, _, ok := boxTriangleSeparation(box, triangle)
	return ok
}

// ResolveBoxTriangle yields a [shape3d.Contact] for the overlap of the box with
// the triangle, if there is one.
//
// The triangle is front-face-culled exactly as in [CheckBoxTriangle], so a box
// centered behind it produces no contact. The contact is reported with the box
// as the source and the triangle as the target. TargetNormal is the
// minimum-translation axis (a unit vector), oriented so that moving the box by
// Depth along it resolves the intersection, and Depth is the overlap along it.
//
// TargetPoint is the spot on the triangle that the penetrating feature (a box
// corner, a triangle vertex, an edge crossing, or a face patch) touches once
// the box is moved out by Depth: the center of the region where the triangle
// meets the box at its just-touching position. It is consistent with
// TargetNormal and Depth, so [shape3d.Contact.EvalSourcePoint] evaluates to
// the matching point on the box surface.
func ResolveBoxTriangle(box shape3d.Box, triangle shape3d.Triangle, yield shape3d.ContactCallback) {
	normal, depth, ok := boxTriangleSeparation(box, triangle)
	if !ok {
		return
	}
	yield(shape3d.Contact{
		TargetPoint:  boxTriangleContactPoint(box, triangle, normal, depth),
		TargetNormal: normal,
		Depth:        depth,
	})
}

// boxTriangleSeparation runs the Separating Axis Theorem for the given box and
// triangle. When they intersect it returns the minimum-translation axis (a
// unit vector oriented so that moving the box along it separates the two
// shapes), the penetration depth along it, and true. When they are disjoint,
// or the box center is not strictly in front of the triangle's plane, it
// returns false.
func boxTriangleSeparation(box shape3d.Box, triangle shape3d.Triangle) (dprec.Vec3, float64, bool) {
	boxAxisX := box.Rotation.BasisX
	boxAxisY := box.Rotation.BasisY
	boxAxisZ := box.Rotation.BasisZ
	center := box.Center

	a := triangle.A
	b := triangle.B
	c := triangle.C

	edgeAB := dprec.Vec3Diff(b, a)
	edgeBC := dprec.Vec3Diff(c, b)
	edgeCA := dprec.Vec3Diff(a, c)
	triangleNormal := dprec.Vec3Cross(edgeAB, dprec.Vec3Diff(c, a))
	if triangleNormal.SqrLength() < boxTriangleAxisEpsilon {
		return dprec.Vec3{}, 0.0, false // degenerate triangle
	}

	// The triangle is one-sided: cull a box whose center is not strictly in
	// front of its plane.
	if dprec.Vec3Dot(triangleNormal, dprec.Vec3Diff(center, a)) <= 0.0 {
		return dprec.Vec3{}, 0.0, false
	}

	axes := [13]dprec.Vec3{
		boxAxisX,
		boxAxisY,
		boxAxisZ,
		triangleNormal,
		dprec.Vec3Cross(boxAxisX, edgeAB),
		dprec.Vec3Cross(boxAxisX, edgeBC),
		dprec.Vec3Cross(boxAxisX, edgeCA),
		dprec.Vec3Cross(boxAxisY, edgeAB),
		dprec.Vec3Cross(boxAxisY, edgeBC),
		dprec.Vec3Cross(boxAxisY, edgeCA),
		dprec.Vec3Cross(boxAxisZ, edgeAB),
		dprec.Vec3Cross(boxAxisZ, edgeBC),
		dprec.Vec3Cross(boxAxisZ, edgeCA),
	}

	bestDepth := math.MaxFloat64
	var bestNormal dprec.Vec3

	for _, axis := range axes {
		axisSqrLength := axis.SqrLength()
		if axisSqrLength < boxTriangleAxisEpsilon {
			continue // degenerate axis from parallel edges
		}

		projA := dprec.Vec3Dot(axis, a)
		projB := dprec.Vec3Dot(axis, b)
		projC := dprec.Vec3Dot(axis, c)
		triangleMin := min(projA, projB, projC)
		triangleMax := max(projA, projB, projC)

		boxReach := box.HalfWidth*dprec.Abs(dprec.Vec3Dot(axis, boxAxisX)) +
			box.HalfHeight*dprec.Abs(dprec.Vec3Dot(axis, boxAxisY)) +
			box.HalfLength*dprec.Abs(dprec.Vec3Dot(axis, boxAxisZ))
		boxCenterProjection := dprec.Vec3Dot(axis, center)
		boxMin := boxCenterProjection - boxReach
		boxMax := boxCenterProjection + boxReach

		// The distances the box must travel along +axis or -axis for the two
		// projected intervals to separate; a negative one means they are already
		// disjoint on that side.
		overlapPositive := triangleMax - boxMin
		overlapNegative := boxMax - triangleMin
		if overlapPositive < 0.0 || overlapNegative < 0.0 {
			return dprec.Vec3{}, 0.0, false // found a separating axis
		}
		overlap, sign := overlapPositive, 1.0
		if overlapNegative < overlapPositive {
			overlap, sign = overlapNegative, -1.0
		}

		// Normalize the overlap to a unit axis so depths are comparable across
		// the differently scaled candidate axes.
		axisLength := dprec.Sqrt(axisSqrLength)
		penetration := overlap / axisLength
		if penetration < bestDepth {
			bestDepth = penetration
			// Orient the axis toward the end whose overlap is smaller, which is
			// the direction the box must move to separate.
			bestNormal = dprec.Vec3Prod(axis, sign/axisLength)
		}
	}

	return bestNormal, bestDepth, true
}

// boxTriangleContactPoint returns the contact point on the triangle for an
// intersection resolved along the given unit normal with the given depth.
//
// The point is the center of the touch region: moving the box by depth along
// the normal brings the two shapes to a just-touching position, and the
// triangle is clipped against the box at that position (inflated slightly for
// robustness). The average of the clipped region's vertices is the contact
// point: the spot the penetrating feature (a box corner, a triangle vertex, an
// edge crossing, or a face patch) touches the triangle at. Because the point
// lies on the translated box, stepping back by depth along the normal, as
// [shape3d.Contact.EvalSourcePoint] does, lands on the original box.
func boxTriangleContactPoint(box shape3d.Box, triangle shape3d.Triangle, normal dprec.Vec3, depth float64) dprec.Vec3 {
	touchCenter := dprec.Vec3Sum(box.Center, dprec.Vec3Prod(normal, depth))
	inflation := 1e-7 * (1.0 + box.HalfWidth + box.HalfHeight + box.HalfLength)

	// A triangle clipped against six half-spaces has at most 3+6 vertices.
	var bufferA, bufferB [9]dprec.Vec3
	polygon := append(bufferA[:0], triangle.A, triangle.B, triangle.C)
	scratch := bufferB[:0]

	slabs := [3]struct {
		axis   dprec.Vec3
		extent float64
	}{
		{box.Rotation.BasisX, box.HalfWidth + inflation},
		{box.Rotation.BasisY, box.HalfHeight + inflation},
		{box.Rotation.BasisZ, box.HalfLength + inflation},
	}
	for _, slab := range slabs {
		centerProjection := dprec.Vec3Dot(slab.axis, touchCenter)
		polygon, scratch = clipPolygonToHalfSpace(polygon, scratch, slab.axis, centerProjection+slab.extent), polygon
		if len(polygon) == 0 {
			break
		}
		polygon, scratch = clipPolygonToHalfSpace(polygon, scratch, dprec.InverseVec3(slab.axis), slab.extent-centerProjection), polygon
		if len(polygon) == 0 {
			break
		}
	}

	if len(polygon) > 0 {
		var sum dprec.Vec3
		for _, vertex := range polygon {
			sum = dprec.Vec3Sum(sum, vertex)
		}
		return dprec.Vec3Quot(sum, float64(len(polygon)))
	}

	// The clip came up empty (a grazing touch lost to floating point): fall
	// back to the deepest point of the box, projected onto the triangle.
	support := boxSupportPoint(box, normal)
	return closestPointOnTriangle(dprec.Vec3Sum(support, dprec.Vec3Prod(normal, depth)), triangle)
}

// clipPolygonToHalfSpace clips the polygon against the half-space of points
// whose projection onto the given axis does not exceed the given limit,
// writing the resulting polygon into dst (which is reset first) and returning
// it.
func clipPolygonToHalfSpace(src, dst []dprec.Vec3, axis dprec.Vec3, limit float64) []dprec.Vec3 {
	dst = dst[:0]
	for i, current := range src {
		next := src[(i+1)%len(src)]
		distCurrent := dprec.Vec3Dot(axis, current) - limit
		distNext := dprec.Vec3Dot(axis, next) - limit
		if distCurrent <= 0.0 {
			dst = append(dst, current)
		}
		if (distCurrent < 0.0 && distNext > 0.0) || (distCurrent > 0.0 && distNext < 0.0) {
			t := distCurrent / (distCurrent - distNext)
			dst = append(dst, dprec.Vec3Lerp(current, next, t))
		}
	}
	return dst
}

// boxSupportPoint returns the point of the box that lies farthest along the
// inverse of the given unit direction. When a box axis is perpendicular to the
// direction the support feature is a face or an edge rather than a corner, and
// the point resolves to that feature's midpoint.
func boxSupportPoint(box shape3d.Box, direction dprec.Vec3) dprec.Vec3 {
	point := box.Center
	point = dprec.Vec3Diff(point, dprec.Vec3Prod(box.Rotation.BasisX, box.HalfWidth*supportSign(dprec.Vec3Dot(direction, box.Rotation.BasisX))))
	point = dprec.Vec3Diff(point, dprec.Vec3Prod(box.Rotation.BasisY, box.HalfHeight*supportSign(dprec.Vec3Dot(direction, box.Rotation.BasisY))))
	point = dprec.Vec3Diff(point, dprec.Vec3Prod(box.Rotation.BasisZ, box.HalfLength*supportSign(dprec.Vec3Dot(direction, box.Rotation.BasisZ))))
	return point
}

// supportSign returns the sign of the given axis projection, treating
// near-perpendicular projections as zero so that flat support features (faces
// and edges) resolve to their midpoint rather than an arbitrary corner.
func supportSign(projection float64) float64 {
	const epsilon = 1e-9
	switch {
	case projection > epsilon:
		return 1.0
	case projection < -epsilon:
		return -1.0
	default:
		return 0.0
	}
}

// closestPointOnTriangle returns the point on the triangle (face, edge, or
// vertex) that is closest to the given point.
func closestPointOnTriangle(point dprec.Vec3, triangle shape3d.Triangle) dprec.Vec3 {
	a := triangle.A
	b := triangle.B
	c := triangle.C

	ab := dprec.Vec3Diff(b, a)
	ac := dprec.Vec3Diff(c, a)
	ap := dprec.Vec3Diff(point, a)

	d1 := dprec.Vec3Dot(ab, ap)
	d2 := dprec.Vec3Dot(ac, ap)
	if d1 <= 0.0 && d2 <= 0.0 {
		return a // vertex A region
	}

	bp := dprec.Vec3Diff(point, b)
	d3 := dprec.Vec3Dot(ab, bp)
	d4 := dprec.Vec3Dot(ac, bp)
	if d3 >= 0.0 && d4 <= d3 {
		return b // vertex B region
	}

	vc := d1*d4 - d3*d2
	if vc <= 0.0 && d1 >= 0.0 && d3 <= 0.0 {
		v := d1 / (d1 - d3)
		return dprec.Vec3Sum(a, dprec.Vec3Prod(ab, v)) // edge AB region
	}

	cp := dprec.Vec3Diff(point, c)
	d5 := dprec.Vec3Dot(ab, cp)
	d6 := dprec.Vec3Dot(ac, cp)
	if d6 >= 0.0 && d5 <= d6 {
		return c // vertex C region
	}

	vb := d5*d2 - d1*d6
	if vb <= 0.0 && d2 >= 0.0 && d6 <= 0.0 {
		w := d2 / (d2 - d6)
		return dprec.Vec3Sum(a, dprec.Vec3Prod(ac, w)) // edge AC region
	}

	va := d3*d6 - d5*d4
	if va <= 0.0 && (d4-d3) >= 0.0 && (d5-d6) >= 0.0 {
		w := (d4 - d3) / ((d4 - d3) + (d5 - d6))
		return dprec.Vec3Sum(b, dprec.Vec3Prod(dprec.Vec3Diff(c, b), w)) // edge BC region
	}

	// Inside the face: convert the barycentric coordinates to a point.
	denom := 1.0 / (va + vb + vc)
	v := vb * denom
	w := vc * denom
	return dprec.Vec3Sum(a, dprec.Vec3Sum(dprec.Vec3Prod(ab, v), dprec.Vec3Prod(ac, w)))
}
