package query3d

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// allFrustumSurfaces is the surface mask with all six surfaces selected.
const allFrustumSurfaces = uint8(0b111111)

// boundingBox is an axis-aligned bounding box, stored through its minimum and
// maximum corner, together with the intersection tests that the query
// structures in this package need. It is shared by [Octree] and [Bag].
type boundingBox struct {
	minX float64
	minY float64
	minZ float64
	maxX float64
	maxY float64
	maxZ float64
}

func emptyBoundingBox() boundingBox {
	return boundingBox{
		minX: math.MaxFloat64,
		minY: math.MaxFloat64,
		minZ: math.MaxFloat64,
		maxX: -math.MaxFloat64,
		maxY: -math.MaxFloat64,
		maxZ: -math.MaxFloat64,
	}
}

func newBoundingBoxFromAABB(aabb shape3d.AABB) boundingBox {
	return boundingBox{
		minX: aabb.MinX,
		minY: aabb.MinY,
		minZ: aabb.MinZ,
		maxX: aabb.MaxX,
		maxY: aabb.MaxY,
		maxZ: aabb.MaxZ,
	}
}

func mergeBoundingBoxes(first, second boundingBox) boundingBox {
	return boundingBox{
		minX: min(first.minX, second.minX),
		minY: min(first.minY, second.minY),
		minZ: min(first.minZ, second.minZ),
		maxX: max(first.maxX, second.maxX),
		maxY: max(first.maxY, second.maxY),
		maxZ: max(first.maxZ, second.maxZ),
	}
}

func (box *boundingBox) isEmpty() bool {
	return (box.minX > box.maxX) || (box.minY > box.maxY) || (box.minZ > box.maxZ)
}

func (box *boundingBox) intersectsSegment(segment *shape3d.Segment) bool {
	if box.isEmpty() {
		return false
	}

	delta := dprec.Vec3Diff(segment.B, segment.A)

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

	var tCloseZ, tFarZ float64
	if delta.Z == 0.0 {
		if (segment.A.Z < box.minZ) || (segment.A.Z > box.maxZ) {
			return false // both points are outside the box on the front or back
		}
		tCloseZ = -math.MaxFloat64
		tFarZ = math.MaxFloat64
	} else {
		tLowZ := (box.minZ - segment.A.Z) / delta.Z
		tHighZ := (box.maxZ - segment.A.Z) / delta.Z
		tCloseZ = min(tLowZ, tHighZ)
		tFarZ = max(tLowZ, tHighZ)
	}

	tClose := max(tCloseX, tCloseY, tCloseZ)
	tFar := min(tFarX, tFarY, tFarZ)

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}

func (box *boundingBox) intersectsAABB(other *shape3d.AABB) bool {
	if box.isEmpty() {
		return false
	}
	return (box.minX <= other.MaxX) &&
		(box.minY <= other.MaxY) &&
		(box.maxX >= other.MinX) &&
		(box.maxY >= other.MinY) &&
		(box.minZ <= other.MaxZ) &&
		(box.maxZ >= other.MinZ)
}

// intersectsFrustum reports whether the box is not fully behind any of the
// surfaces of the frustum that are selected by the mask. For each surface,
// only the box corner that lies farthest along the surface normal is tested:
// if even that corner is behind the surface, the whole box is.
func (box *boundingBox) intersectsFrustum(frustum *shape3d.Frustum, mask uint8) bool {
	if box.isEmpty() {
		return false
	}
	for i := range frustum.Surfaces {
		if mask&(1<<i) == 0 {
			continue
		}
		surface := &frustum.Surfaces[i]
		farPoint := dprec.NewVec3(box.maxX, box.maxY, box.maxZ)
		if surface.Normal.X < 0.0 {
			farPoint.X = box.minX
		}
		if surface.Normal.Y < 0.0 {
			farPoint.Y = box.minY
		}
		if surface.Normal.Z < 0.0 {
			farPoint.Z = box.minZ
		}
		if dprec.Vec3Dot(surface.Normal, farPoint) < surface.Distance {
			return false
		}
	}
	return true
}

// classifyFrustum is like intersectsFrustum but additionally returns
// the subset of the mask for which the box is not fully in front of the
// surface. Surfaces that the box is fully in front of need not be tested for
// anything contained within the box.
func (box *boundingBox) classifyFrustum(frustum *shape3d.Frustum, mask uint8) (uint8, bool) {
	if box.isEmpty() {
		return mask, false
	}
	for i := range frustum.Surfaces {
		if mask&(1<<i) == 0 {
			continue
		}
		surface := &frustum.Surfaces[i]
		nearPoint := dprec.NewVec3(box.minX, box.minY, box.minZ)
		farPoint := dprec.NewVec3(box.maxX, box.maxY, box.maxZ)
		if surface.Normal.X < 0.0 {
			nearPoint.X, farPoint.X = farPoint.X, nearPoint.X
		}
		if surface.Normal.Y < 0.0 {
			nearPoint.Y, farPoint.Y = farPoint.Y, nearPoint.Y
		}
		if surface.Normal.Z < 0.0 {
			nearPoint.Z, farPoint.Z = farPoint.Z, nearPoint.Z
		}
		if dprec.Vec3Dot(surface.Normal, farPoint) < surface.Distance {
			return mask, false // fully behind the surface
		}
		if dprec.Vec3Dot(surface.Normal, nearPoint) >= surface.Distance {
			mask &^= 1 << i // fully in front of the surface
		}
	}
	return mask, true
}
