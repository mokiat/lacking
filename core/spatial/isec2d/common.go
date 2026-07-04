package isec2d

import "math"

func slabRange(start, delta, halfExtent float64) (float64, float64, bool) {
	if delta == 0 {
		if (start < -halfExtent) || (start > halfExtent) {
			return 0.0, 0.0, false // both points are outside the box on the left or right
		}
		return -math.MaxFloat64, math.MaxFloat64, true
	}
	tLow := (-halfExtent - start) / delta
	tHigh := (halfExtent - start) / delta
	if tLow < tHigh {
		return tLow, tHigh, true
	} else {
		return tHigh, tLow, true
	}
}
