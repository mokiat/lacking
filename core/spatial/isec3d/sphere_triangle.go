package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSphereTriangle reports whether the sphere intersects the triangle.
//
// The triangle is treated as a one-sided, front-face-culled surface: only a
// sphere whose center lies strictly in front of the triangle (on the side its
// normal faces) and within the sphere radius of the triangle's plane can
// intersect it. A sphere centered behind the triangle, or exactly on its plane,
// is never considered to intersect it, even if it would otherwise overlap. A
// sphere whose center is in front and that merely touches the triangle, whether
// on its face, along an edge, or at a vertex, counts as intersecting.
func CheckSphereTriangle(sphere shape3d.Sphere, triangle shape3d.Triangle) bool {
	spherePosition := sphere.Center
	sphereRadius := sphere.Radius
	triangleA := triangle.A
	triangleB := triangle.B
	triangleC := triangle.C
	triangleNormal := triangle.Normal()

	sphereOffset := dprec.Vec3Diff(spherePosition, triangleA)
	height := dprec.Vec3Dot(triangleNormal, sphereOffset)
	if height > sphereRadius || height <= 0 {
		return false
	}

	vecAB := dprec.Vec3Diff(triangleB, triangleA)
	vecBC := dprec.Vec3Diff(triangleC, triangleB)
	vecCA := dprec.Vec3Diff(triangleA, triangleC)
	tangentAB := dprec.UnitVec3(vecAB)
	tangentBC := dprec.UnitVec3(vecBC)
	tangentCA := dprec.UnitVec3(vecCA)
	normAB := dprec.Vec3Cross(tangentAB, triangleNormal)
	normBC := dprec.Vec3Cross(tangentBC, triangleNormal)
	normCA := dprec.Vec3Cross(tangentCA, triangleNormal)

	projectedPoint := dprec.Vec3Diff(spherePosition, dprec.Vec3Prod(triangleNormal, height))
	vecAP := dprec.Vec3Diff(projectedPoint, triangleA)
	vecBP := dprec.Vec3Diff(projectedPoint, triangleB)
	vecCP := dprec.Vec3Diff(projectedPoint, triangleC)

	distAB := dprec.Vec3Dot(normAB, vecAP)
	distBC := dprec.Vec3Dot(normBC, vecBP)
	distCA := dprec.Vec3Dot(normCA, vecCP)

	var (
		outsideAB bool
		outsideBC bool
		outsideCA bool
		outsideA  bool
		outsideB  bool
		outsideC  bool
	)
	switch {
	case distAB >= 0:
		if dprec.Vec3Dot(tangentAB, vecAP) >= 0 {
			if dprec.Vec3Dot(tangentAB, vecBP) <= 0 {
				outsideAB = true
			} else {
				outsideB = true
			}
		} else {
			outsideA = true
		}
	case distBC >= 0:
		if dprec.Vec3Dot(tangentBC, vecBP) >= 0 {
			if dprec.Vec3Dot(tangentBC, vecCP) <= 0 {
				outsideBC = true
			} else {
				outsideC = true
			}
		} else {
			outsideB = true
		}
	case distCA >= 0:
		if dprec.Vec3Dot(tangentCA, vecCP) >= 0 {
			if dprec.Vec3Dot(tangentCA, vecAP) <= 0 {
				outsideCA = true
			} else {
				outsideA = true
			}
		} else {
			outsideC = true
		}
	}

	switch {
	case outsideA:
		cornerOffset := dprec.Vec3Diff(spherePosition, triangleA)
		cornerDistance := cornerOffset.Length()
		return cornerDistance <= sphereRadius

	case outsideB:
		cornerOffset := dprec.Vec3Diff(spherePosition, triangleB)
		cornerDistance := cornerOffset.Length()
		return cornerDistance <= sphereRadius

	case outsideC:
		cornerOffset := dprec.Vec3Diff(spherePosition, triangleC)
		cornerDistance := cornerOffset.Length()
		return cornerDistance <= sphereRadius

	case outsideAB:
		edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normAB, distAB), dprec.Vec3Prod(triangleNormal, height))
		edgeDistance := edgeOffset.Length()
		return edgeDistance <= sphereRadius

	case outsideBC:
		edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normBC, distBC), dprec.Vec3Prod(triangleNormal, height))
		edgeDistance := edgeOffset.Length()
		return edgeDistance <= sphereRadius

	case outsideCA:
		edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normCA, distCA), dprec.Vec3Prod(triangleNormal, height))
		edgeDistance := edgeOffset.Length()
		return edgeDistance <= sphereRadius

	default: // inside
		return true
	}
}

// ResolveSphereTriangle yields a [shape3d.Contact] for the overlap of the
// sphere with the triangle, if there is one.
//
// The triangle is front-face-culled exactly as in [CheckSphereTriangle], so a
// sphere centered behind it produces no contact. The contact is reported with
// the sphere as the source and the triangle as the target: TargetPoint is the
// point on the triangle closest to the sphere center (on the face, an edge, or a
// vertex), TargetNormal is the outward unit direction from that point toward the
// sphere center, and Depth is how far the sphere penetrates, that is the sphere
// radius minus the distance to that closest point. Moving the sphere by Depth
// along TargetNormal resolves the overlap.
func ResolveSphereTriangle(sphere shape3d.Sphere, triangle shape3d.Triangle, yield shape3d.ContactCallback) {
	spherePosition := sphere.Center
	sphereRadius := sphere.Radius
	triangleA := triangle.A
	triangleB := triangle.B
	triangleC := triangle.C
	triangleNormal := triangle.Normal()

	sphereOffset := dprec.Vec3Diff(spherePosition, triangleA)
	height := dprec.Vec3Dot(triangleNormal, sphereOffset)
	if height > sphereRadius || height <= 0 {
		return
	}

	vecAB := dprec.Vec3Diff(triangleB, triangleA)
	vecBC := dprec.Vec3Diff(triangleC, triangleB)
	vecCA := dprec.Vec3Diff(triangleA, triangleC)
	tangentAB := dprec.UnitVec3(vecAB)
	tangentBC := dprec.UnitVec3(vecBC)
	tangentCA := dprec.UnitVec3(vecCA)
	normAB := dprec.Vec3Cross(tangentAB, triangleNormal)
	normBC := dprec.Vec3Cross(tangentBC, triangleNormal)
	normCA := dprec.Vec3Cross(tangentCA, triangleNormal)

	projectedPoint := dprec.Vec3Diff(spherePosition, dprec.Vec3Prod(triangleNormal, height))
	vecAP := dprec.Vec3Diff(projectedPoint, triangleA)
	vecBP := dprec.Vec3Diff(projectedPoint, triangleB)
	vecCP := dprec.Vec3Diff(projectedPoint, triangleC)

	distAB := dprec.Vec3Dot(normAB, vecAP)
	distBC := dprec.Vec3Dot(normBC, vecBP)
	distCA := dprec.Vec3Dot(normCA, vecCP)

	var (
		outsideAB bool
		outsideBC bool
		outsideCA bool
		outsideA  bool
		outsideB  bool
		outsideC  bool
	)
	switch {
	case distAB >= 0:
		if dprec.Vec3Dot(tangentAB, vecAP) >= 0 {
			if dprec.Vec3Dot(tangentAB, vecBP) <= 0 {
				outsideAB = true
			} else {
				outsideB = true
			}
		} else {
			outsideA = true
		}
	case distBC >= 0:
		if dprec.Vec3Dot(tangentBC, vecBP) >= 0 {
			if dprec.Vec3Dot(tangentBC, vecCP) <= 0 {
				outsideBC = true
			} else {
				outsideC = true
			}
		} else {
			outsideB = true
		}
	case distCA >= 0:
		if dprec.Vec3Dot(tangentCA, vecCP) >= 0 {
			if dprec.Vec3Dot(tangentCA, vecAP) <= 0 {
				outsideCA = true
			} else {
				outsideA = true
			}
		} else {
			outsideC = true
		}
	}

	var (
		isIntersection       bool
		depth                float64
		sphereDisplaceNormal dprec.Vec3
	)
	switch {
	case outsideA:
		cornerOffset := dprec.Vec3Diff(spherePosition, triangleA)
		cornerDistance := cornerOffset.Length()
		if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
			depth = sphereRadius - cornerDistance
			sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
		}

	case outsideB:
		cornerOffset := dprec.Vec3Diff(spherePosition, triangleB)
		cornerDistance := cornerOffset.Length()
		if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
			depth = sphereRadius - cornerDistance
			sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
		}

	case outsideC:
		cornerOffset := dprec.Vec3Diff(spherePosition, triangleC)
		cornerDistance := cornerOffset.Length()
		if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
			depth = sphereRadius - cornerDistance
			sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
		}

	case outsideAB:
		edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normAB, distAB), dprec.Vec3Prod(triangleNormal, height))
		edgeDistance := edgeOffset.Length()
		if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
			depth = sphereRadius - edgeDistance
			sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
		}

	case outsideBC:
		edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normBC, distBC), dprec.Vec3Prod(triangleNormal, height))
		edgeDistance := edgeOffset.Length()
		if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
			depth = sphereRadius - edgeDistance
			sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
		}

	case outsideCA:
		edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normCA, distCA), dprec.Vec3Prod(triangleNormal, height))
		edgeDistance := edgeOffset.Length()
		if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
			depth = sphereRadius - edgeDistance
			sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
		}

	default: // inside
		isIntersection = true
		depth = sphereRadius - height
		sphereDisplaceNormal = triangleNormal
	}

	if isIntersection {
		yield(shape3d.Contact{
			TargetPoint:  dprec.Vec3Diff(spherePosition, dprec.Vec3Prod(sphereDisplaceNormal, sphereRadius-depth)),
			TargetNormal: sphereDisplaceNormal,
			Depth:        depth,
		})
	}
}
