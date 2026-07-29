package query3d

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// InvalidTreeItemID is an identifier that can be used by user
// code to mark an identifier as invalid. Such an identifier will
// never be returned by the library but must also never be passed to the
// library.
const InvalidTreeItemID = TreeItemID(0xFFFFFFFF)

// TreeItemID is an identifier used to control the placement of an item
// into a tree.
type TreeItemID uint32

// TreeStats represents the current state of a tree.
type TreeStats struct {

	// NodeCount is the total number of nodes in the tree.
	NodeCount uint32

	// ItemCount is the total number of items in the tree.
	ItemCount uint32

	// ItemCountPerDepth contains the number of items at each depth level.
	ItemCountPerDepth []uint32
}

// TreeVisitStats represents statistics on the last visit operation
// performed on a tree.
type TreeVisitStats struct {

	// NodeCountVisited is the number of nodes that were visited during the last
	// visit operation.
	NodeCountVisited uint32

	// NodeCountAccepted is the number of nodes that were determined relevant
	// during the last visit operation.
	NodeCountAccepted uint32

	// NodeCountRejected is the number of nodes that were determined irrelevant
	// during the last visit operation.
	NodeCountRejected uint32

	// ItemCountVisited is the number of items that were visited during the last
	// visit operation.
	ItemCountVisited uint32

	// ItemCountAccepted is the number of items that were determined relevant
	// during the last visit operation.
	ItemCountAccepted uint32

	// ItemCountRejected is the number of items that were determined irrelevant
	// during the last visit operation.
	ItemCountRejected uint32
}

// treeAABB is the internal axis-aligned bounding box that the trees in this
// package use to store the extent of their items and nodes.
//
// It is shared by all tree implementations so that they all reach the same
// accept and reject decisions for a given item and query.
type treeAABB struct {
	minX float64
	minY float64
	minZ float64
	maxX float64
	maxY float64
	maxZ float64
}

func emptyTreeAABB() treeAABB {
	return treeAABB{
		minX: math.MaxFloat64,
		minY: math.MaxFloat64,
		minZ: math.MaxFloat64,
		maxX: -math.MaxFloat64,
		maxY: -math.MaxFloat64,
		maxZ: -math.MaxFloat64,
	}
}

func newTreeAABBFromArea(area Area) treeAABB {
	return treeAABB{
		minX: area.x - area.r,
		minY: area.y - area.r,
		minZ: area.z - area.r,
		maxX: area.x + area.r,
		maxY: area.y + area.r,
		maxZ: area.z + area.r,
	}
}

// newFatTreeAABBFromArea creates a [treeAABB] that covers the specified area
// grown by the specified margin on all sides.
func newFatTreeAABBFromArea(area Area, margin float64) treeAABB {
	radius := area.r + margin
	return treeAABB{
		minX: area.x - radius,
		minY: area.y - radius,
		minZ: area.z - radius,
		maxX: area.x + radius,
		maxY: area.y + radius,
		maxZ: area.z + radius,
	}
}

func mergeTreeAABBs(first, second treeAABB) treeAABB {
	return treeAABB{
		minX: min(first.minX, second.minX),
		minY: min(first.minY, second.minY),
		minZ: min(first.minZ, second.minZ),
		maxX: max(first.maxX, second.maxX),
		maxY: max(first.maxY, second.maxY),
		maxZ: max(first.maxZ, second.maxZ),
	}
}

func (aabb *treeAABB) isEmpty() bool {
	return (aabb.minX > aabb.maxX) || (aabb.minY > aabb.maxY) || (aabb.minZ > aabb.maxZ)
}

// merged returns the smallest box that contains both this box and other.
//
// It is a pointer based variant of [mergeTreeAABBs] that avoids copying the
// operands, which matters on the hot paths of a [BVHTree].
func (aabb *treeAABB) merged(other *treeAABB) treeAABB {
	return treeAABB{
		minX: min(aabb.minX, other.minX),
		minY: min(aabb.minY, other.minY),
		minZ: min(aabb.minZ, other.minZ),
		maxX: max(aabb.maxX, other.maxX),
		maxY: max(aabb.maxY, other.maxY),
		maxZ: max(aabb.maxZ, other.maxZ),
	}
}

// area returns a value that is proportional to the surface area of this box.
//
// The constant factor of two that a true surface area would have is omitted,
// since all cost comparisons that make use of this value scale by it uniformly.
func (aabb *treeAABB) area() float64 {
	deltaX := aabb.maxX - aabb.minX
	deltaY := aabb.maxY - aabb.minY
	deltaZ := aabb.maxZ - aabb.minZ
	return deltaX*deltaY + deltaY*deltaZ + deltaZ*deltaX
}

// contains returns whether the specified box lies fully inside this box.
func (aabb *treeAABB) contains(other *treeAABB) bool {
	return (aabb.minX <= other.minX) && (aabb.maxX >= other.maxX) &&
		(aabb.minY <= other.minY) && (aabb.maxY >= other.maxY) &&
		(aabb.minZ <= other.minZ) && (aabb.maxZ >= other.maxZ)
}

func (aabb *treeAABB) intersectsSegment(segment *Segment) bool {
	if aabb.isEmpty() {
		return false
	}

	delta := dprec.Vec3Diff(segment.b, segment.a)

	var tCloseX, tFarX float64
	if delta.X == 0.0 {
		if (segment.a.X < aabb.minX) || (segment.a.X > aabb.maxX) {
			return false // both points are outside the box on the left or right
		}
		tCloseX = -math.MaxFloat64
		tFarX = math.MaxFloat64
	} else {
		tLowX := (aabb.minX - segment.a.X) / delta.X
		tHighX := (aabb.maxX - segment.a.X) / delta.X
		tCloseX = min(tLowX, tHighX)
		tFarX = max(tLowX, tHighX)
	}

	var tCloseY, tFarY float64
	if delta.Y == 0.0 {
		if (segment.a.Y < aabb.minY) || (segment.a.Y > aabb.maxY) {
			return false // both points are outside the box on the top or bottom
		}
		tCloseY = -math.MaxFloat64
		tFarY = math.MaxFloat64
	} else {
		tLowY := (aabb.minY - segment.a.Y) / delta.Y
		tHighY := (aabb.maxY - segment.a.Y) / delta.Y
		tCloseY = min(tLowY, tHighY)
		tFarY = max(tLowY, tHighY)
	}

	var tCloseZ, tFarZ float64
	if delta.Z == 0.0 {
		if (segment.a.Z < aabb.minZ) || (segment.a.Z > aabb.maxZ) {
			return false // both points are outside the box on the front or back
		}
		tCloseZ = -math.MaxFloat64
		tFarZ = math.MaxFloat64
	} else {
		tLowZ := (aabb.minZ - segment.a.Z) / delta.Z
		tHighZ := (aabb.maxZ - segment.a.Z) / delta.Z
		tCloseZ = min(tLowZ, tHighZ)
		tFarZ = max(tLowZ, tHighZ)
	}

	tClose := max(tCloseX, tCloseY, tCloseZ)
	tFar := min(tFarX, tFarY, tFarZ)

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}

func (aabb *treeAABB) intersectsAABB(other *AABB) bool {
	if aabb.isEmpty() {
		return false
	}
	return (aabb.minX <= other.maxX) &&
		(aabb.minY <= other.maxY) &&
		(aabb.maxX >= other.minX) &&
		(aabb.maxY >= other.minY) &&
		(aabb.minZ <= other.maxZ) &&
		(aabb.maxZ >= other.minZ)
}

// overlapsAABB is a variant of [treeAABB.intersectsAABB] that omits the check
// for an empty box. It must only be used on boxes that are known to be
// non-empty, which is the case for all nodes of a [BVHTree].
//
// On a non-empty box it evaluates the exact same predicate, since it only
// performs comparisons and no arithmetic.
func (aabb *treeAABB) overlapsAABB(other *AABB) bool {
	return (aabb.minX <= other.maxX) &&
		(aabb.minY <= other.maxY) &&
		(aabb.maxX >= other.minX) &&
		(aabb.maxY >= other.minY) &&
		(aabb.minZ <= other.maxZ) &&
		(aabb.maxZ >= other.minZ)
}

// overlapsRay is a variant of [treeAABB.intersectsSegment] that uses a
// precomputed [bvhRay] in order to replace the six divisions of the slab test
// with six multiplications.
//
// It must only be used on boxes that are known to be non-empty, which is the
// case for all nodes of a [BVHTree]. Since a reciprocal introduces a rounding
// error of at most a couple of units in the last place, the outcome can differ
// from [treeAABB.intersectsSegment] for segments that graze the box. This is
// acceptable for node boxes, which are conservative supersets of the item
// boxes and are re-tested exactly once traversal reaches a leaf.
func (aabb *treeAABB) overlapsRay(ray *bvhRay) bool {
	tLowX := (aabb.minX - ray.originX) * ray.invDeltaX
	tHighX := (aabb.maxX - ray.originX) * ray.invDeltaX
	tLowY := (aabb.minY - ray.originY) * ray.invDeltaY
	tHighY := (aabb.maxY - ray.originY) * ray.invDeltaY
	tLowZ := (aabb.minZ - ray.originZ) * ray.invDeltaZ
	tHighZ := (aabb.maxZ - ray.originZ) * ray.invDeltaZ

	tClose := max(min(tLowX, tHighX), min(tLowY, tHighY), min(tLowZ, tHighZ))
	tFar := min(max(tLowX, tHighX), max(tLowY, tHighY), max(tLowZ, tHighZ))

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}
