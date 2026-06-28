package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSphereBox reports whether the sphere and the box intersect.
//
// A sphere that merely touches the box, or that lies entirely inside it, is
// considered to intersect. The box may be arbitrarily oriented; the test is
// performed in the box's local frame.
func CheckSphereBox(sphere shape3d.Sphere, box shape3d.Box) bool {
	spherePosition := sphere.Center
	sphereRadius := sphere.Radius

	boxPosition := box.Center
	boxRotation := box.Rotation
	boxAxisX := boxRotation.BasisX
	boxAxisY := boxRotation.BasisY
	boxAxisZ := boxRotation.BasisZ
	boxHalfWidth := box.HalfWidth
	boxHalfHeight := box.HalfHeight
	boxHalfLength := box.HalfLength

	deltaPosition := dprec.Vec3Diff(spherePosition, boxPosition)
	distanceX := dprec.Vec3Dot(boxAxisX, deltaPosition)
	distanceY := dprec.Vec3Dot(boxAxisY, deltaPosition)
	distanceZ := dprec.Vec3Dot(boxAxisZ, deltaPosition)

	distanceRight := distanceX - boxHalfWidth
	if distanceRight > sphereRadius {
		return false
	}

	distanceLeft := -distanceX - boxHalfWidth
	if distanceLeft > sphereRadius {
		return false
	}

	distanceTop := distanceY - boxHalfHeight
	if distanceTop > sphereRadius {
		return false
	}

	distanceBottom := -distanceY - boxHalfHeight
	if distanceBottom > sphereRadius {
		return false
	}

	distanceFront := distanceZ - boxHalfLength
	if distanceFront > sphereRadius {
		return false
	}

	distanceBack := -distanceZ - boxHalfLength
	if distanceBack > sphereRadius {
		return false
	}

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
	case maskLeft, maskRight, maskBottom, maskTop, maskBack, maskFront:
		return true

	case maskLeft | maskBottom:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom
		return sqrDistance <= sphereRadius*sphereRadius

	case maskLeft | maskTop:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop
		return sqrDistance <= sphereRadius*sphereRadius

	case maskRight | maskBottom:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom
		return sqrDistance <= sphereRadius*sphereRadius

	case maskRight | maskTop:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop
		return sqrDistance <= sphereRadius*sphereRadius

	case maskBack | maskBottom:
		sqrDistance := distanceBack*distanceBack + distanceBottom*distanceBottom
		return sqrDistance <= sphereRadius*sphereRadius

	case maskBack | maskTop:
		sqrDistance := distanceBack*distanceBack + distanceTop*distanceTop
		return sqrDistance <= sphereRadius*sphereRadius

	case maskFront | maskBottom:
		sqrDistance := distanceFront*distanceFront + distanceBottom*distanceBottom
		return sqrDistance <= sphereRadius*sphereRadius

	case maskFront | maskTop:
		sqrDistance := distanceFront*distanceFront + distanceTop*distanceTop
		return sqrDistance <= sphereRadius*sphereRadius

	case maskBack | maskLeft:
		sqrDistance := distanceBack*distanceBack + distanceLeft*distanceLeft
		return sqrDistance <= sphereRadius*sphereRadius

	case maskBack | maskRight:
		sqrDistance := distanceBack*distanceBack + distanceRight*distanceRight
		return sqrDistance <= sphereRadius*sphereRadius

	case maskFront | maskLeft:
		sqrDistance := distanceFront*distanceFront + distanceLeft*distanceLeft
		return sqrDistance <= sphereRadius*sphereRadius

	case maskFront | maskRight:
		sqrDistance := distanceFront*distanceFront + distanceRight*distanceRight
		return sqrDistance <= sphereRadius*sphereRadius

	case maskLeft | maskBottom | maskBack:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom + distanceBack*distanceBack
		return sqrDistance <= sphereRadius*sphereRadius

	case maskLeft | maskBottom | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom + distanceFront*distanceFront
		return sqrDistance <= sphereRadius*sphereRadius

	case maskLeft | maskTop | maskBack:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceBack*distanceBack
		return sqrDistance <= sphereRadius*sphereRadius

	case maskLeft | maskTop | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceFront*distanceFront
		return sqrDistance <= sphereRadius*sphereRadius

	case maskRight | maskBottom | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceBack*distanceBack
		return sqrDistance <= sphereRadius*sphereRadius

	case maskRight | maskBottom | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceFront*distanceFront
		return sqrDistance <= sphereRadius*sphereRadius

	case maskRight | maskTop | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceBack*distanceBack
		return sqrDistance <= sphereRadius*sphereRadius

	case maskRight | maskTop | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceFront*distanceFront
		return sqrDistance <= sphereRadius*sphereRadius

	default: // inside box
		return true
	}
}

// ResolveSphereBox yields a [shape3d.Contact] for the overlap of the sphere
// with the box, if there is one.
//
// The contact is reported with the sphere as the source and the box as the
// target: TargetPoint is the point of the box closest to the sphere center
// (lying on the box surface), TargetNormal is the outward direction from that
// point toward the sphere center, and Depth is how far the sphere penetrates the
// box along that normal. Moving the sphere by Depth along TargetNormal resolves
// the overlap.
//
// When the sphere center lies inside the box there is no well-defined closest
// surface point, so the normal is chosen along the box axis of least
// penetration and the depth is measured from the corresponding face.
func ResolveSphereBox(sphere shape3d.Sphere, box shape3d.Box, yield shape3d.ContactCallback) {
	spherePosition := sphere.Center
	sphereRadius := sphere.Radius

	boxPosition := box.Center
	boxRotation := box.Rotation
	boxAxisX := boxRotation.BasisX
	boxAxisY := boxRotation.BasisY
	boxAxisZ := boxRotation.BasisZ
	boxHalfWidth := box.HalfWidth
	boxHalfHeight := box.HalfHeight
	boxHalfLength := box.HalfLength

	deltaPosition := dprec.Vec3Diff(spherePosition, boxPosition)
	distanceX := dprec.Vec3Dot(boxAxisX, deltaPosition)
	distanceY := dprec.Vec3Dot(boxAxisY, deltaPosition)
	distanceZ := dprec.Vec3Dot(boxAxisZ, deltaPosition)

	distanceRight := distanceX - boxHalfWidth
	if distanceRight > sphereRadius {
		return
	}

	distanceLeft := -distanceX - boxHalfWidth
	if distanceLeft > sphereRadius {
		return
	}

	distanceTop := distanceY - boxHalfHeight
	if distanceTop > sphereRadius {
		return
	}

	distanceBottom := -distanceY - boxHalfHeight
	if distanceBottom > sphereRadius {
		return
	}

	distanceFront := distanceZ - boxHalfLength
	if distanceFront > sphereRadius {
		return
	}

	distanceBack := -distanceZ - boxHalfLength
	if distanceBack > sphereRadius {
		return
	}

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

	var (
		isIntersection    bool
		depth             float64
		boxContact        dprec.Vec3
		boxDisplaceNormal dprec.Vec3
	)

	switch mask {
	case maskLeft:
		isIntersection = true
		depth = sphereRadius - distanceLeft
		boxDisplaceNormal = boxAxisX
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))

	case maskRight:
		isIntersection = true
		depth = sphereRadius - distanceRight
		boxDisplaceNormal = dprec.InverseVec3(boxAxisX)
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))

	case maskBottom:
		isIntersection = true
		depth = sphereRadius - distanceBottom
		boxDisplaceNormal = boxAxisY
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))

	case maskTop:
		isIntersection = true
		depth = sphereRadius - distanceTop
		boxDisplaceNormal = dprec.InverseVec3(boxAxisY)
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))

	case maskBack:
		isIntersection = true
		depth = sphereRadius - distanceBack
		boxDisplaceNormal = boxAxisZ
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))

	case maskFront:
		isIntersection = true
		depth = sphereRadius - distanceFront
		boxDisplaceNormal = dprec.InverseVec3(boxAxisZ)
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))

	case maskLeft | maskBottom:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
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
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskBottom | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom + distanceFront*distanceFront
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskTop | maskBack:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceBack*distanceBack
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskTop | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceFront*distanceFront
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskBottom | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceBack*distanceBack
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskBottom | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceFront*distanceFront
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskTop | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceBack*distanceBack
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskTop | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceFront*distanceFront
		if isIntersection = sqrDistance <= sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	default: // inside box
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

	if isIntersection {
		yield(shape3d.Contact{
			TargetPoint:  boxContact,
			TargetNormal: dprec.InverseVec3(boxDisplaceNormal),
			Depth:        depth,
		})
	}
}
