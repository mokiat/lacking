package shape2d

import "github.com/mokiat/gomath/dprec"

// CheckCircleRectangleIntersection checks if a circle shape intersects with
// a rectangle shape.
func CheckCircleRectangleIntersection(circle Circle, rectangle Rectangle, yield IntersectionYieldFunc) {
	circlePosition := circle.Position
	circleRadius := circle.Radius

	rectanglePosition := rectangle.Position
	rectangleRotation := rectangle.Rotation
	rectangleAxisX := dprec.AngleVec2Rotation(rectangleRotation, dprec.BasisXVec2())
	rectangleAxisY := dprec.AngleVec2Rotation(rectangleRotation, dprec.BasisYVec2())
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
		if isIntersection = sqrDistance < circleRadius*circleRadius; isIntersection {
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
		if isIntersection = sqrDistance < circleRadius*circleRadius; isIntersection {
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
		if isIntersection = sqrDistance < circleRadius*circleRadius; isIntersection {
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
		if isIntersection = sqrDistance < circleRadius*circleRadius; isIntersection {
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
		yield(Intersection{
			TargetContact: targetContact,
			TargetNormal:  targetNormal,
			Depth:         depth,
		})
	}
}
