package isec2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// CheckCircleRectangle reports whether the circle intersects the rectangle.
//
// A circle that merely touches the rectangle, or that lies entirely inside it,
// is considered to intersect. The rectangle may be arbitrarily oriented; the
// test is performed in the rectangle's local frame.
func CheckCircleRectangle(circle shape2d.Circle, rectangle shape2d.Rectangle) bool {
	circlePosition := circle.Center
	circleRadius := circle.Radius

	rectanglePosition := rectangle.Center
	rectangleRotation := rectangle.Rotation
	rectangleAxisX := rectangleRotation.BasisX
	rectangleAxisY := rectangleRotation.BasisY
	rectangleHalfWidth := rectangle.HalfWidth
	rectangleHalfHeight := rectangle.HalfHeight

	deltaPosition := dprec.Vec2Diff(circlePosition, rectanglePosition)
	distanceX := dprec.Vec2Dot(deltaPosition, rectangleAxisX)
	distanceY := dprec.Vec2Dot(deltaPosition, rectangleAxisY)

	distanceRight := distanceX - rectangleHalfWidth
	if distanceRight > circleRadius {
		return false
	}

	distanceLeft := -distanceX - rectangleHalfWidth
	if distanceLeft > circleRadius {
		return false
	}

	distanceTop := distanceY - rectangleHalfHeight
	if distanceTop > circleRadius {
		return false
	}

	distanceBottom := -distanceY - rectangleHalfHeight
	if distanceBottom > circleRadius {
		return false
	}

	const (
		maskLeft   = 0b1000
		maskRight  = 0b0100
		maskBottom = 0b0010
		maskTop    = 0b0001
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

	switch mask {
	case maskLeft, maskRight, maskBottom, maskTop:
		return true

	case maskLeft | maskBottom:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom
		return sqrDistance <= circleRadius*circleRadius

	case maskLeft | maskTop:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop
		return sqrDistance <= circleRadius*circleRadius

	case maskRight | maskBottom:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom
		return sqrDistance <= circleRadius*circleRadius

	case maskRight | maskTop:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop
		return sqrDistance <= circleRadius*circleRadius

	default: // inside rectangle
		return true
	}
}

// ResolveCircleRectangle yields a [shape2d.Contact] for the overlap of the
// circle with the rectangle, if there is one.
//
// The contact is reported with the circle as the source and the rectangle as the
// target: TargetPoint is the point of the rectangle closest to the circle center
// (lying on the rectangle's perimeter), TargetNormal is the outward direction
// from that point toward the circle center, and Depth is how far the circle
// penetrates the rectangle along that normal. Moving the circle by Depth along
// TargetNormal resolves the overlap.
//
// When the circle center lies inside the rectangle there is no closest perimeter
// direction, so the contact is resolved along the rectangle axis of least
// penetration.
func ResolveCircleRectangle(circle shape2d.Circle, rectangle shape2d.Rectangle, yield shape2d.ContactCallback) {
	circlePosition := circle.Center
	circleRadius := circle.Radius

	rectanglePosition := rectangle.Center
	rectangleRotation := rectangle.Rotation
	rectangleAxisX := rectangleRotation.BasisX
	rectangleAxisY := rectangleRotation.BasisY
	rectangleHalfWidth := rectangle.HalfWidth
	rectangleHalfHeight := rectangle.HalfHeight

	deltaPosition := dprec.Vec2Diff(circlePosition, rectanglePosition)
	distanceX := dprec.Vec2Dot(deltaPosition, rectangleAxisX)
	distanceY := dprec.Vec2Dot(deltaPosition, rectangleAxisY)

	distanceRight := distanceX - rectangleHalfWidth
	if distanceRight > circleRadius {
		return
	}

	distanceLeft := -distanceX - rectangleHalfWidth
	if distanceLeft > circleRadius {
		return
	}

	distanceTop := distanceY - rectangleHalfHeight
	if distanceTop > circleRadius {
		return
	}

	distanceBottom := -distanceY - rectangleHalfHeight
	if distanceBottom > circleRadius {
		return
	}

	const (
		maskLeft   = 0b1000
		maskRight  = 0b0100
		maskBottom = 0b0010
		maskTop    = 0b0001
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

	var (
		isIntersection bool
		depth          float64
		targetContact  dprec.Vec2
		targetNormal   dprec.Vec2
	)

	switch mask {
	case maskLeft:
		isIntersection = true
		depth = circleRadius - distanceLeft
		targetNormal = dprec.InverseVec2(rectangleAxisX)
		targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distanceLeft))

	case maskRight:
		isIntersection = true
		depth = circleRadius - distanceRight
		targetNormal = rectangleAxisX
		targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distanceRight))

	case maskBottom:
		isIntersection = true
		depth = circleRadius - distanceBottom
		targetNormal = dprec.InverseVec2(rectangleAxisY)
		targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distanceBottom))

	case maskTop:
		isIntersection = true
		depth = circleRadius - distanceTop
		targetNormal = rectangleAxisY
		targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distanceTop))

	case maskLeft | maskBottom:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom
		if isIntersection = sqrDistance <= circleRadius*circleRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = circleRadius - distance
			targetNormal = dprec.UnitVec2(dprec.Vec2Sum(
				dprec.Vec2Prod(rectangleAxisX, -distanceLeft),
				dprec.Vec2Prod(rectangleAxisY, -distanceBottom),
			))
			targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distance))
		}

	case maskLeft | maskTop:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop
		if isIntersection = sqrDistance <= circleRadius*circleRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = circleRadius - distance
			targetNormal = dprec.UnitVec2(dprec.Vec2Sum(
				dprec.Vec2Prod(rectangleAxisX, -distanceLeft),
				dprec.Vec2Prod(rectangleAxisY, distanceTop),
			))
			targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distance))
		}

	case maskRight | maskBottom:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom
		if isIntersection = sqrDistance <= circleRadius*circleRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = circleRadius - distance
			targetNormal = dprec.UnitVec2(dprec.Vec2Sum(
				dprec.Vec2Prod(rectangleAxisX, distanceRight),
				dprec.Vec2Prod(rectangleAxisY, -distanceBottom),
			))
			targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distance))
		}

	case maskRight | maskTop:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop
		if isIntersection = sqrDistance <= circleRadius*circleRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = circleRadius - distance
			targetNormal = dprec.UnitVec2(dprec.Vec2Sum(
				dprec.Vec2Prod(rectangleAxisX, distanceRight),
				dprec.Vec2Prod(rectangleAxisY, distanceTop),
			))
			targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, -distance))
		}

	default: // inside rectangle
		isIntersection = true
		var (
			displaceX float64
			displaceY float64
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
		if dprec.Abs(displaceX) < dprec.Abs(displaceY) {
			depth = dprec.Abs(displaceX) + circleRadius
			targetNormal = dprec.Vec2Prod(rectangleAxisX, dprec.Sign(displaceX))
		} else {
			depth = dprec.Abs(displaceY) + circleRadius
			targetNormal = dprec.Vec2Prod(rectangleAxisY, dprec.Sign(displaceY))
		}
		targetContact = dprec.Vec2Sum(circlePosition, dprec.Vec2Prod(targetNormal, depth-circleRadius))
	}

	if isIntersection {
		yield(shape2d.Contact{
			TargetPoint:  targetContact,
			TargetNormal: targetNormal,
			Depth:        depth,
		})
	}
}
