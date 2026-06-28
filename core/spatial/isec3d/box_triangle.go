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
// The test is the Separating Axis Theorem applied to the box's three face
// normals, the triangle's normal, and the nine box-edge-cross-triangle-edge
// axes. Unlike [CheckSphereTriangle], it is two-sided: the triangle is treated
// as a flat solid that the box can touch from either face. A box that merely
// touches the triangle counts as intersecting.
func CheckBoxTriangle(box shape3d.Box, triangle shape3d.Triangle) bool {
	_, _, ok := boxTriangleSeparation(box, triangle)
	return ok
}

// ResolveBoxTriangle yields a [shape3d.Contact] for the overlap of the box with
// the triangle, if there is one.
//
// The contact is reported with the box as the source and the triangle as the
// target. TargetNormal is the minimum-translation axis (a unit vector) that
// separates the two shapes, oriented to point from the triangle toward the box
// center, and Depth is the overlap along it, so moving the box by Depth along
// TargetNormal resolves the intersection. TargetPoint is the point on the
// triangle closest to the box center, which serves as a representative contact
// location. As with [CheckBoxTriangle], the test is two-sided.
func ResolveBoxTriangle(box shape3d.Box, triangle shape3d.Triangle, yield shape3d.ContactCallback) {
	normal, depth, ok := boxTriangleSeparation(box, triangle)
	if !ok {
		return
	}
	yield(shape3d.Contact{
		TargetPoint:  closestPointOnTriangle(box.Center, triangle),
		TargetNormal: normal,
		Depth:        depth,
	})
}

// boxTriangleSeparation runs the Separating Axis Theorem for the given box and
// triangle. When they intersect it returns the minimum-translation axis (a unit
// vector oriented from the triangle toward the box center), the penetration
// depth along it, and true. When they are disjoint it returns false.
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
	centroid := triangle.Centroid()

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

		// The minimum translation that separates the two projected intervals,
		// which also correctly reports the penetration when one interval (the
		// thin triangle) lies inside the other.
		overlap := min(boxMax-triangleMin, triangleMax-boxMin)
		if overlap < 0.0 {
			return dprec.Vec3{}, 0.0, false // found a separating axis
		}

		// Normalize the overlap to a unit axis so depths are comparable across
		// the differently scaled candidate axes.
		axisLength := dprec.Sqrt(axisSqrLength)
		penetration := overlap / axisLength
		if penetration < bestDepth {
			bestDepth = penetration
			unitAxis := dprec.Vec3Quot(axis, axisLength)
			// Orient the axis so it points from the triangle toward the box
			// center, i.e. the direction the box must move to separate.
			if dprec.Vec3Dot(unitAxis, dprec.Vec3Diff(center, centroid)) < 0.0 {
				unitAxis = dprec.InverseVec3(unitAxis)
			}
			bestNormal = unitAxis
		}
	}

	return bestNormal, bestDepth, true
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
