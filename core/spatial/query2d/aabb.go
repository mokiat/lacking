package query2d

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// allFrustumSurfaces is the surface mask with all four surfaces selected.
const allFrustumSurfaces = uint8(0b1111)

// boundingBox is an axis-aligned bounding box, stored through its minimum and
// maximum corner, together with the intersection tests that the query
// structures in this package need. It is shared by [Quadtree] and [Bag].
type boundingBox struct {
	minX float64
	minY float64
	maxX float64
	maxY float64
}

func emptyBoundingBox() boundingBox {
	return boundingBox{
		minX: math.MaxFloat64,
		minY: math.MaxFloat64,
		maxX: -math.MaxFloat64,
		maxY: -math.MaxFloat64,
	}
}

func newBoundingBoxFromAABB(aabb shape2d.AABB) boundingBox {
	return boundingBox{
		minX: aabb.MinX,
		minY: aabb.MinY,
		maxX: aabb.MaxX,
		maxY: aabb.MaxY,
	}
}

func mergeBoundingBoxes(first, second boundingBox) boundingBox {
	return boundingBox{
		minX: min(first.minX, second.minX),
		minY: min(first.minY, second.minY),
		maxX: max(first.maxX, second.maxX),
		maxY: max(first.maxY, second.maxY),
	}
}

func (box *boundingBox) isEmpty() bool {
	return (box.minX > box.maxX) || (box.minY > box.maxY)
}

func (box *boundingBox) intersectsSegment(segment *shape2d.Segment) bool {
	if box.isEmpty() {
		return false
	}

	delta := dprec.Vec2Diff(segment.B, segment.A)

	var tCloseX, tFarX float64
	if delta.X == 0.0 {
		if (segment.A.X < box.minX) || (segment.A.X > box.maxX) {
			return false // both points are outside the box on the left or right
		}
		tCloseX = -math.MaxFloat64
		tFarX = math.MaxFloat64
	} else {
		tLowX := (box.minX - segment.A.X) / delta.X
		tHighX := (box.maxX - segment.A.X) / delta.X
		tCloseX = min(tLowX, tHighX)
		tFarX = max(tLowX, tHighX)
	}

	var tCloseY, tFarY float64
	if delta.Y == 0.0 {
		if (segment.A.Y < box.minY) || (segment.A.Y > box.maxY) {
			return false // both points are outside the box on the top or bottom
		}
		tCloseY = -math.MaxFloat64
		tFarY = math.MaxFloat64
	} else {
		tLowY := (box.minY - segment.A.Y) / delta.Y
		tHighY := (box.maxY - segment.A.Y) / delta.Y
		tCloseY = min(tLowY, tHighY)
		tFarY = max(tLowY, tHighY)
	}

	tClose := max(tCloseX, tCloseY)
	tFar := min(tFarX, tFarY)

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}

func (box *boundingBox) intersectsAABB(other *shape2d.AABB) bool {
	if box.isEmpty() {
		return false
	}
	return (box.minX <= other.MaxX) &&
		(box.minY <= other.MaxY) &&
		(box.maxX >= other.MinX) &&
		(box.maxY >= other.MinY)
}

// intersectsFrustum reports whether the box is not fully behind any of the
// surfaces of the frustum that are selected by the mask. For each surface,
// only the box corner that lies farthest along the surface normal is tested:
// if even that corner is behind the surface, the whole box is.
func (box *boundingBox) intersectsFrustum(frustum *shape2d.Frustum, mask uint8) bool {
	if box.isEmpty() {
		return false
	}
	for i := range frustum.Surfaces {
		if mask&(1<<i) == 0 {
			continue
		}
		surface := &frustum.Surfaces[i]
		farPoint := dprec.NewVec2(box.maxX, box.maxY)
		if surface.Normal.X < 0.0 {
			farPoint.X = box.minX
		}
		if surface.Normal.Y < 0.0 {
			farPoint.Y = box.minY
		}
		if dprec.Vec2Dot(surface.Normal, farPoint) < surface.Distance {
			return false
		}
	}
	return true
}

// classifyFrustum is like intersectsFrustum but additionally returns
// the subset of the mask for which the box is not fully in front of the
// surface. Surfaces that the box is fully in front of need not be tested for
// anything contained within the box.
func (box *boundingBox) classifyFrustum(frustum *shape2d.Frustum, mask uint8) (uint8, bool) {
	if box.isEmpty() {
		return mask, false
	}
	for i := range frustum.Surfaces {
		if mask&(1<<i) == 0 {
			continue
		}
		surface := &frustum.Surfaces[i]
		nearPoint := dprec.NewVec2(box.minX, box.minY)
		farPoint := dprec.NewVec2(box.maxX, box.maxY)
		if surface.Normal.X < 0.0 {
			nearPoint.X, farPoint.X = farPoint.X, nearPoint.X
		}
		if surface.Normal.Y < 0.0 {
			nearPoint.Y, farPoint.Y = farPoint.Y, nearPoint.Y
		}
		if dprec.Vec2Dot(surface.Normal, farPoint) < surface.Distance {
			return mask, false // fully behind the surface
		}
		if dprec.Vec2Dot(surface.Normal, nearPoint) >= surface.Distance {
			mask &^= 1 << i // fully in front of the surface
		}
	}
	return mask, true
}
