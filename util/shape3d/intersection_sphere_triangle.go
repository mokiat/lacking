package shape3d

import (
	"github.com/mokiat/gomath/dprec"
)

// TODO: Revisit this!
func CheckSphereTriangleIntersection(sphere Sphere, triangle Triangle) (Intersection, bool) {
	spherePosition := sphere.Position
	sphereRadius := sphere.Radius
	triangleA := triangle.A
	triangleB := triangle.B
	triangleC := triangle.C
	triangleCenter := triangle.Center()
	triangleNormal := triangle.Normal()

	sphereOffset := dprec.Vec3Diff(spherePosition, triangleCenter)
	height := dprec.Vec3Dot(triangleNormal, sphereOffset)
	if height > sphereRadius || height < 0 {
		return Intersection{}, false
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
		inside    bool
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
	default:
		inside = true
	}

	var (
		isIntersection       bool
		depth                float64
		sphereDisplaceNormal dprec.Vec3
	)
	switch {
	// TODO: Recover all cases once the physics engine is fixed to check
	// collisions precisely (via binary search or similar).

	case outsideA:
	// 	cornerOffset := dprec.Vec3Diff(spherePosition, triangleA)
	// 	cornerDistance := cornerOffset.Length()
	// 	if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - cornerDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
	// 	}

	case outsideB:
	// 	cornerOffset := dprec.Vec3Diff(spherePosition, triangleB)
	// 	cornerDistance := cornerOffset.Length()
	// 	if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - cornerDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
	// 	}

	case outsideC:
	// 	cornerOffset := dprec.Vec3Diff(spherePosition, triangleC)
	// 	cornerDistance := cornerOffset.Length()
	// 	if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - cornerDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
	// 	}

	case outsideAB:
	// 	edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normAB, distAB), dprec.Vec3Prod(triangleNormal, height))
	// 	edgeDistance := edgeOffset.Length()
	// 	if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - edgeDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
	// 	}

	case outsideBC:
	// 	edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normBC, distBC), dprec.Vec3Prod(triangleNormal, height))
	// 	edgeDistance := edgeOffset.Length()
	// 	if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - edgeDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
	// 	}

	case outsideCA:
	// 	edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normCA, distCA), dprec.Vec3Prod(triangleNormal, height))
	// 	edgeDistance := edgeOffset.Length()
	// 	if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - edgeDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
	// 	}

	case inside:
		isIntersection = true
		depth = sphereRadius - height
		sphereDisplaceNormal = triangleNormal
	}

	if !isIntersection {
		return Intersection{}, false
	}

	return Intersection{
		TargetContact: dprec.Vec3Diff(spherePosition, dprec.Vec3Prod(sphereDisplaceNormal, sphereRadius-depth)),
		TargetNormal:  sphereDisplaceNormal,
		Depth:         depth,
	}, true
}
