package shape3d

import "github.com/mokiat/gomath/dprec"

// CheckSphereBoxIntersection checks if the specified sphere intersects with
// the specified box.
func CheckSphereBoxIntersection(sphere Sphere, box Box) (Intersection, bool) {
	spherePosition := sphere.Position
	sphereRadius := sphere.Radius

	boxPosition := box.Position
	boxRotation := box.Rotation
	boxAxisX := boxRotation.OrientationX()
	boxAxisY := boxRotation.OrientationY()
	boxAxisZ := boxRotation.OrientationZ()
	boxHalfWidth := box.HalfWidth
	boxHalfHeight := box.HalfHeight
	boxHalfLength := box.HalfLength

	deltaPosition := dprec.Vec3Diff(spherePosition, boxPosition)
	distanceX := dprec.Vec3Dot(boxAxisX, deltaPosition)
	distanceY := dprec.Vec3Dot(boxAxisY, deltaPosition)
	distanceZ := dprec.Vec3Dot(boxAxisZ, deltaPosition)

	distanceRight := distanceX - boxHalfWidth
	distanceLeft := -(2.0*boxHalfWidth + distanceRight)
	distanceTop := distanceY - boxHalfHeight
	distanceBottom := -(2.0*boxHalfHeight + distanceTop)
	distanceFront := distanceZ - boxHalfLength
	distanceBack := -(2.0*boxHalfLength + distanceFront)

	var (
		isIntersection    bool
		depth             float64
		boxContact        dprec.Vec3
		boxDisplaceNormal dprec.Vec3
	)

	const (
		maskLeft   = 0b100000
		maskRight  = 0b010000
		maskBottom = 0b001000
		maskTop    = 0b000100
		maskBack   = 0b000010
		maskFront  = 0b000001
	)
	var mask uint8
	if distanceLeft > 0 {
		mask |= maskLeft
	}
	if distanceRight > 0 {
		mask |= maskRight
	}
	if distanceBottom > 0 {
		mask |= maskBottom
	}
	if distanceTop > 0 {
		mask |= maskTop
	}
	if distanceBack > 0 {
		mask |= maskBack
	}
	if distanceFront > 0 {
		mask |= maskFront
	}
	switch mask {
	case maskLeft:
		if depth = sphereRadius - distanceLeft; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = boxAxisX
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskRight:
		if depth = sphereRadius - distanceRight; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = dprec.InverseVec3(boxAxisX)
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskBottom:
		if depth = sphereRadius - distanceBottom; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = boxAxisY
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskTop:
		if depth = sphereRadius - distanceTop; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = dprec.InverseVec3(boxAxisY)
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskBack:
		if depth = sphereRadius - distanceBack; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = boxAxisZ
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskFront:
		if depth = sphereRadius - distanceFront; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = dprec.InverseVec3(boxAxisZ)
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskLeft | maskBottom:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisY, distanceBottom),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskLeft | maskTop:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisY, -distanceTop),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskRight | maskBottom:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisY, distanceBottom),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskRight | maskTop:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisY, -distanceTop),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskBottom:
		sqrDistance := distanceBack*distanceBack + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, distanceBottom),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskTop:
		sqrDistance := distanceBack*distanceBack + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, -distanceTop),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskBottom:
		sqrDistance := distanceFront*distanceFront + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, distanceBottom),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskTop:
		sqrDistance := distanceFront*distanceFront + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, -distanceTop),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskLeft:
		sqrDistance := distanceBack*distanceBack + distanceLeft*distanceLeft
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskRight:
		sqrDistance := distanceBack*distanceBack + distanceRight*distanceRight
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskLeft:
		sqrDistance := distanceFront*distanceFront + distanceLeft*distanceLeft
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskRight:
		sqrDistance := distanceFront*distanceFront + distanceRight*distanceRight
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskLeft | maskBottom | maskBack:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskBottom | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskTop | maskBack:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskTop | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskBottom | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskBottom | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskTop | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskTop | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	default:
		// Note: This branch is unlikely to occur so no need to be extremely optimal.
		isIntersection = true
		var (
			displaceX float64
			displaceY float64
			displaceZ float64
		)
		if distanceLeft > distanceRight {
			displaceX = distanceLeft
		} else {
			displaceX = -distanceRight
		}
		if distanceBottom > distanceTop {
			displaceY = distanceBottom
		} else {
			displaceY = -distanceTop
		}
		if distanceBack > distanceFront {
			displaceZ = distanceBack
		} else {
			displaceZ = -distanceFront
		}
		if dprec.Abs(displaceX) < dprec.Abs(displaceY) {
			if dprec.Abs(displaceX) < dprec.Abs(displaceZ) {
				depth = dprec.Abs(displaceX) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisX, -dprec.Sign(displaceX))
			} else {
				depth = dprec.Abs(displaceZ) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisZ, -dprec.Sign(displaceZ))
			}
		} else {
			if dprec.Abs(displaceY) < dprec.Abs(displaceZ) {
				depth = dprec.Abs(displaceY) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisY, -dprec.Sign(displaceY))
			} else {
				depth = dprec.Abs(displaceZ) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisZ, -dprec.Sign(displaceZ))
			}
		}
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
	}

	if !isIntersection {
		return Intersection{}, false
	}

	return Intersection{
		TargetContact: boxContact,
		TargetNormal:  boxDisplaceNormal,
		Depth:         depth,
	}, true
}
