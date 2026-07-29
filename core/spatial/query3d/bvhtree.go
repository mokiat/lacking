package query3d

import (
	"math"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
)

// BVHTreeSettings contains the settings for a [BVHTree].
type BVHTreeSettings struct {

	// AABBMargin is the distance by which the bounding box of an item is grown
	// when it is placed into the tree. An item that moves within its grown box
	// does not require any change to the structure of the tree, which makes
	// [BVHTree.Update] very cheap.
	//
	// A larger margin makes updates cheaper but makes queries slightly less
	// precise. A good value is roughly the distance that a typical item travels
	// between two updates.
	//
	// If not specified, a default margin of 0.1 is used. Regardless of this
	// setting, the margin is never smaller than a small fraction of the radius
	// of the item, so that it remains meaningful for very large items.
	AABBMargin opt.T[float64]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. A tree that holds N items uses exactly 2*N-1
	// nodes, so this should normally be twice InitialItemCapacity. This allows
	// the tree to preallocate memory and avoid dynamic allocations.
	//
	// By default the initial capacity is 2048.
	InitialNodeCapacity opt.T[uint32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the tree. This allows the tree to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[uint32]
}

// BVHTree is a spatial structure that uses a dynamic bounding volume hierarchy
// to enable the fast searching of items.
//
// It is a drop-in replacement for an [Octree] and returns the exact same items
// for a given query. Unlike an [Octree], it has neither fixed bounds nor a
// maximum depth, and it reorganizes itself as items are added so that the
// bounding boxes of sibling nodes overlap as little as possible. This makes it
// a better fit for scenes that are large, sparse or unbounded, and for scenes
// in which items move frequently, at the cost of a more expensive insertion.
type BVHTree[T any] struct {
	nodes           []bvhNode
	items           []bvhItem[T]
	freeNodeIndices *ds.Stack[int32]
	freeItemIDs     *ds.Stack[TreeItemID]
	root            int32
	aabbMargin      float64

	nodeCountAccepted uint32
	nodeCountRejected uint32
	itemCountAccepted uint32
	itemCountRejected uint32
}

// NewBVHTree creates a new [BVHTree] using the provided settings.
func NewBVHTree[T any](settings BVHTreeSettings) *BVHTree[T] {
	aabbMargin := settings.AABBMargin.ValueOrDefault(0.1)
	if aabbMargin < 0.0 {
		panic("aabb margin cannot be negative")
	}
	initialNodeCapacity := settings.InitialNodeCapacity.ValueOrDefault(2048)
	initialItemCapacity := settings.InitialItemCapacity.ValueOrDefault(1024)

	return &BVHTree[T]{
		nodes:           make([]bvhNode, 0, initialNodeCapacity),
		items:           make([]bvhItem[T], 0, initialItemCapacity),
		freeNodeIndices: ds.EmptyStack[int32](),
		freeItemIDs:     ds.EmptyStack[TreeItemID](),
		root:            nullBVHIndex,
		aabbMargin:      aabbMargin,

		nodeCountAccepted: 0,
		nodeCountRejected: 0,
		itemCountAccepted: 0,
		itemCountRejected: 0,
	}
}

// Stats returns statistics on the current state of this tree.
//
// Unlike an [Octree], which always has a root node, an empty [BVHTree] has no
// nodes at all. Furthermore, the length of ItemCountPerDepth reflects the
// current height of the tree and hence changes as items are added and removed.
func (t *BVHTree[T]) Stats() TreeStats {
	var itemCountPerDepth []uint32
	if t.root != nullBVHIndex {
		itemCountPerDepth = make([]uint32, t.nodes[t.root].height+1)
		t.countItemsPerDepth(t.root, 0, itemCountPerDepth)
	}
	return TreeStats{
		NodeCount:         t.activeNodeCount(),
		ItemCount:         t.activeItemCount(),
		ItemCountPerDepth: itemCountPerDepth,
	}
}

// VisitStats returns statistics information on the last executed search in
// this tree.
func (t *BVHTree[T]) VisitStats() TreeVisitStats {
	return TreeVisitStats{
		NodeCountVisited:  t.nodeCountAccepted + t.nodeCountRejected,
		NodeCountAccepted: t.nodeCountAccepted,
		NodeCountRejected: t.nodeCountRejected,
		ItemCountVisited:  t.itemCountAccepted + t.itemCountRejected,
		ItemCountAccepted: t.itemCountAccepted,
		ItemCountRejected: t.itemCountRejected,
	}
}

// Insert adds an item, which occupies the specified area, to this tree.
func (t *BVHTree[T]) Insert(area Area, value T) TreeItemID {
	box := newTreeAABBFromArea(area)
	fatBox := newFatTreeAABBFromArea(area, t.marginFor(area))

	nodeIndex := t.allocateNode()

	var id TreeItemID
	if t.freeItemIDs.IsEmpty() {
		id = TreeItemID(len(t.items))
		t.items = append(t.items, bvhItem[T]{
			box:   box,
			node:  nodeIndex,
			value: value,
		})
	} else {
		id = t.freeItemIDs.Pop()
		item := &t.items[id]
		item.box = box
		item.node = nodeIndex
		item.value = value
	}

	// Do NOT acquire a node reference before the allocation above, as the
	// nodes slice might have been replaced by a larger one in the meantime.
	node := &t.nodes[nodeIndex]
	node.box = fatBox
	node.parent = nullBVHIndex
	node.childLeft = nullBVHIndex
	node.childRight = int32(id)
	node.height = 0

	t.insertLeaf(nodeIndex)
	return id
}

// Update repositions the item with the specified id to the new area.
//
// As long as the new area remains within the grown bounding box that was
// computed when the item was last placed, this is a very cheap operation that
// leaves the structure of the tree untouched.
func (t *BVHTree[T]) Update(id TreeItemID, area Area) {
	item := &t.items[id]
	if item.node == nullBVHIndex {
		panic("cannot update removed item")
	}
	box := newTreeAABBFromArea(area)
	item.box = box

	nodeIndex := item.node
	if t.nodes[nodeIndex].box.contains(&box) {
		return // the grown box of the leaf still covers the item
	}

	// The leaf node and the item slot are reused, so the id remains valid.
	t.removeLeaf(nodeIndex)
	node := &t.nodes[nodeIndex]
	node.box = newFatTreeAABBFromArea(area, t.marginFor(area))
	node.parent = nullBVHIndex
	node.childLeft = nullBVHIndex
	node.childRight = int32(id)
	node.height = 0
	t.insertLeaf(nodeIndex)
}

// Remove removes the item with the specified id from this tree.
func (t *BVHTree[T]) Remove(id TreeItemID) {
	item := &t.items[id]
	if item.node == nullBVHIndex {
		panic("cannot remove item twice")
	}
	nodeIndex := item.node
	item.node = nullBVHIndex

	item.value = gog.Zero[T]() // avoid keeping the old value alive
	t.freeItemIDs.Push(id)

	t.removeLeaf(nodeIndex)
	t.releaseNode(nodeIndex)
}

// QuerySegment finds all items that intersect the specified segment. Each
// found item is passed to the specified yield function. The order in which
// items are passed is undefined and might change between invocations.
func (t *BVHTree[T]) QuerySegment(segment Segment, yield VisitorFunc[T]) {
	t.resetVisitStats()
	if t.root == nullBVHIndex {
		return
	}
	ray := newBVHRay(&segment)

	// A balanced tree never grows deep enough to exhaust this buffer, so the
	// traversal performs no allocations. The buffer is local to the call so
	// that a query issued from within the yield function still works.
	var buffer [bvhStackCapacity]int32
	stack := append(buffer[:0], t.root)

	for len(stack) > 0 {
		nodeIndex := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		node := &t.nodes[nodeIndex]
		if !node.box.overlapsRay(&ray) {
			t.nodeCountRejected++
			continue
		}
		t.nodeCountAccepted++

		if node.isLeaf() {
			item := &t.items[node.childRight]
			if item.box.intersectsSegment(&segment) {
				t.itemCountAccepted++
				if !yield(item.value) {
					return
				}
			} else {
				t.itemCountRejected++
			}
			continue
		}
		// This appends to the traversal stack and not to the nodes slice, so
		// the node reference above remains valid.
		stack = append(stack, node.childLeft, node.childRight)
	}
}

// QueryAABB finds all items that are inside or intersect the specified
// axis-aligned bounding box. Each found item is passed to the specified yield
// function. The order in which items are passed is undefined and might change
// between invocations.
func (t *BVHTree[T]) QueryAABB(aabb AABB, yield VisitorFunc[T]) {
	t.resetVisitStats()
	if t.root == nullBVHIndex {
		return
	}

	// A balanced tree never grows deep enough to exhaust this buffer, so the
	// traversal performs no allocations. The buffer is local to the call so
	// that a query issued from within the yield function still works.
	var buffer [bvhStackCapacity]int32
	stack := append(buffer[:0], t.root)

	for len(stack) > 0 {
		nodeIndex := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		node := &t.nodes[nodeIndex]
		if !node.box.overlapsAABB(&aabb) {
			t.nodeCountRejected++
			continue
		}
		t.nodeCountAccepted++

		if node.isLeaf() {
			item := &t.items[node.childRight]
			if item.box.intersectsAABB(&aabb) {
				t.itemCountAccepted++
				if !yield(item.value) {
					return
				}
			} else {
				t.itemCountRejected++
			}
			continue
		}
		// This appends to the traversal stack and not to the nodes slice, so
		// the node reference above remains valid.
		stack = append(stack, node.childLeft, node.childRight)
	}
}

func (t *BVHTree[T]) resetVisitStats() {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
}

func (t *BVHTree[T]) activeNodeCount() uint32 {
	return uint32(len(t.nodes) - t.freeNodeIndices.Size())
}

func (t *BVHTree[T]) activeItemCount() uint32 {
	return uint32(len(t.items) - t.freeItemIDs.Size())
}

func (t *BVHTree[T]) countItemsPerDepth(nodeIndex int32, depth uint32, counts []uint32) {
	node := &t.nodes[nodeIndex]
	if node.isLeaf() {
		counts[depth]++
		return
	}
	t.countItemsPerDepth(node.childLeft, depth+1, counts)
	t.countItemsPerDepth(node.childRight, depth+1, counts)
}

// marginFor returns the distance by which the bounding box of the specified
// area should be grown.
//
// The configured margin acts as an absolute lower bound, which keeps the
// margin useful for point-like items, whereas the relative term keeps it
// useful for very large items in scenes that span multiple orders of
// magnitude.
func (t *BVHTree[T]) marginFor(area Area) float64 {
	return max(t.aabbMargin, area.r*bvhRelativeMargin)
}

func (t *BVHTree[T]) allocateNode() int32 {
	if t.freeNodeIndices.IsEmpty() {
		nodeIndex := int32(len(t.nodes))
		// Do NOT hold a node reference across this append, as the slice might
		// be replaced by a larger one.
		t.nodes = append(t.nodes, bvhNode{})
		return nodeIndex
	}
	return t.freeNodeIndices.Pop()
}

func (t *BVHTree[T]) releaseNode(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	node.parent = nullBVHIndex
	node.childLeft = nullBVHIndex
	node.childRight = nullBVHIndex
	node.height = freeBVHNodeHeight
	t.freeNodeIndices.Push(nodeIndex)
}

// insertLeaf attaches the specified leaf node to the tree at the location that
// results in the smallest increase of the total surface area of the tree.
func (t *BVHTree[T]) insertLeaf(leafIndex int32) {
	if t.root == nullBVHIndex {
		t.root = leafIndex
		t.nodes[leafIndex].parent = nullBVHIndex
		return
	}
	leafBox := t.nodes[leafIndex].box

	// Find the best sibling for the new leaf by descending the tree for as
	// long as pushing the leaf further down remains cheaper than stopping.
	siblingIndex := t.root
	for !t.nodes[siblingIndex].isLeaf() {
		node := &t.nodes[siblingIndex]
		leftIndex := node.childLeft
		rightIndex := node.childRight

		nodeArea := node.box.area()
		combinedBox := node.box.merged(&leafBox)
		combinedArea := combinedBox.area()

		// The cost of creating a new parent for this node and the new leaf.
		costHere := 2.0 * combinedArea
		// The minimum cost that pushing the leaf further down would add to
		// every node along the way.
		inheritanceCost := 2.0 * (combinedArea - nodeArea)

		costLeft := t.descentCost(leftIndex, &leafBox, inheritanceCost)
		costRight := t.descentCost(rightIndex, &leafBox, inheritanceCost)

		if (costHere < costLeft) && (costHere < costRight) {
			break
		}
		if costLeft < costRight {
			siblingIndex = leftIndex
		} else {
			siblingIndex = rightIndex
		}
	}

	// Insert a new parent node above the chosen sibling.
	oldParentIndex := t.nodes[siblingIndex].parent
	newParentIndex := t.allocateNode()

	newParent := &t.nodes[newParentIndex]
	newParent.parent = oldParentIndex
	newParent.childLeft = siblingIndex
	newParent.childRight = leafIndex
	newParent.box = t.nodes[siblingIndex].box.merged(&leafBox)
	newParent.height = t.nodes[siblingIndex].height + 1

	t.nodes[siblingIndex].parent = newParentIndex
	t.nodes[leafIndex].parent = newParentIndex

	if oldParentIndex == nullBVHIndex {
		t.root = newParentIndex
	} else {
		oldParent := &t.nodes[oldParentIndex]
		if oldParent.childLeft == siblingIndex {
			oldParent.childLeft = newParentIndex
		} else {
			oldParent.childRight = newParentIndex
		}
	}

	t.refitAncestors(newParentIndex, withRotation)
}

// descentCost returns the cost of pushing a leaf with the specified box into
// the subtree that is rooted at the specified node.
func (t *BVHTree[T]) descentCost(nodeIndex int32, leafBox *treeAABB, inheritanceCost float64) float64 {
	node := &t.nodes[nodeIndex]
	combinedBox := node.box.merged(leafBox)
	if node.isLeaf() {
		return combinedBox.area() + inheritanceCost
	}
	return (combinedBox.area() - node.box.area()) + inheritanceCost
}

// removeLeaf detaches the specified leaf node from the tree. The leaf node
// itself remains allocated, so that the caller can either reuse or release it.
func (t *BVHTree[T]) removeLeaf(leafIndex int32) {
	if leafIndex == t.root {
		t.root = nullBVHIndex
		t.nodes[leafIndex].parent = nullBVHIndex
		return
	}

	parentIndex := t.nodes[leafIndex].parent
	parent := &t.nodes[parentIndex]
	grandParentIndex := parent.parent

	siblingIndex := parent.childLeft
	if siblingIndex == leafIndex {
		siblingIndex = parent.childRight
	}

	// The parent node becomes redundant, so the sibling takes its place.
	t.nodes[leafIndex].parent = nullBVHIndex
	t.nodes[siblingIndex].parent = grandParentIndex

	if grandParentIndex == nullBVHIndex {
		t.root = siblingIndex
		t.releaseNode(parentIndex)
		return
	}

	grandParent := &t.nodes[grandParentIndex]
	if grandParent.childLeft == parentIndex {
		grandParent.childLeft = siblingIndex
	} else {
		grandParent.childRight = siblingIndex
	}
	t.releaseNode(parentIndex)

	t.refitAncestors(grandParentIndex, withoutRotation)
}

// Values for the rotate parameter of [BVHTree.refitAncestors].
const (
	withRotation    = true
	withoutRotation = false
)

// refitAncestors walks from the specified node up towards the root, optionally
// improving each subtree along the way and recomputing its bounding box and
// height.
//
// The caller guarantees that the children of the first node have changed. The
// walk stops as soon as a node turns out to be unaffected, since in that case
// none of its ancestors can be affected either. On a large tree this is what
// keeps an insertion or a removal from touching every level.
//
// Rotations are worth their cost only when the tree grows. A removal can only
// shrink the boxes of the nodes that it touches, so the quality of the tree
// is preserved by the rotations that the following insertions perform anyway.
func (t *BVHTree[T]) refitAncestors(nodeIndex int32, rotate bool) {
	for isFirst := true; nodeIndex != nullBVHIndex; isFirst = false {
		node := &t.nodes[nodeIndex]
		oldBox := node.box
		oldHeight := node.height

		t.refreshNode(nodeIndex)
		if !isFirst && (node.box == oldBox) && (node.height == oldHeight) {
			return
		}

		if rotate {
			t.rotate(nodeIndex)
			// A rotation preserves the set of items below the node and hence
			// also its box, but it can change its height.
			node.height = 1 + max(
				t.nodes[node.childLeft].height,
				t.nodes[node.childRight].height,
			)
		}

		nodeIndex = node.parent
	}
}

// refreshNode recomputes the bounding box and the height of the specified
// node from its two children.
func (t *BVHTree[T]) refreshNode(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	left := &t.nodes[node.childLeft]
	right := &t.nodes[node.childRight]
	node.height = 1 + max(left.height, right.height)
	node.box = left.box.merged(&right.box)
}

// rotate improves the subtree that is rooted at the specified node by swapping
// one of its children with one of its grandchildren, whenever doing so shrinks
// the bounding box of the affected child.
//
// Only the four swaps that leave the box of the node itself unchanged are
// considered, which makes the operation strictly local. Rotating towards a
// smaller total surface area, rather than towards an even height, is what
// keeps queries fast: a height driven rotation regularly forces spatially
// distant items to share a node and roughly doubles the number of nodes that
// a query has to visit.
func (t *BVHTree[T]) rotate(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	if node.isLeaf() {
		return
	}
	leftIndex := node.childLeft
	rightIndex := node.childRight
	left := &t.nodes[leftIndex]
	right := &t.nodes[rightIndex]

	var (
		bestGain       float64 // a swap must strictly improve on this
		bestChild      = nullBVHIndex
		bestGrandChild = nullBVHIndex
	)

	if !right.isLeaf() {
		// Move the left child down into the right child, in exchange for one
		// of the children of the right child.
		currentArea := right.box.area()
		rightLeft := &t.nodes[right.childLeft]
		rightRight := &t.nodes[right.childRight]

		box := left.box.merged(&rightRight.box)
		if gain := currentArea - box.area(); gain > bestGain {
			bestGain = gain
			bestChild = leftIndex
			bestGrandChild = right.childLeft
		}
		box = rightLeft.box.merged(&left.box)
		if gain := currentArea - box.area(); gain > bestGain {
			bestGain = gain
			bestChild = leftIndex
			bestGrandChild = right.childRight
		}
	}

	if !left.isLeaf() {
		// The mirrored case: move the right child down into the left child.
		currentArea := left.box.area()
		leftLeft := &t.nodes[left.childLeft]
		leftRight := &t.nodes[left.childRight]

		box := right.box.merged(&leftRight.box)
		if gain := currentArea - box.area(); gain > bestGain {
			bestGain = gain
			bestChild = rightIndex
			bestGrandChild = left.childLeft
		}
		box = leftLeft.box.merged(&right.box)
		if gain := currentArea - box.area(); gain > bestGain {
			bestGain = gain
			bestChild = rightIndex
			bestGrandChild = left.childRight
		}
	}

	if bestGain <= 0.0 {
		return // none of the swaps is an improvement
	}
	t.swapNodes(bestChild, bestGrandChild)
}

// swapNodes exchanges the positions of a child of a node with one of the
// grandchildren of that same node.
func (t *BVHTree[T]) swapNodes(childIndex, grandChildIndex int32) {
	child := &t.nodes[childIndex]
	grandChild := &t.nodes[grandChildIndex]

	parentIndex := child.parent
	siblingIndex := grandChild.parent

	parent := &t.nodes[parentIndex]
	if parent.childLeft == childIndex {
		parent.childLeft = grandChildIndex
	} else {
		parent.childRight = grandChildIndex
	}
	sibling := &t.nodes[siblingIndex]
	if sibling.childLeft == grandChildIndex {
		sibling.childLeft = childIndex
	} else {
		sibling.childRight = childIndex
	}
	child.parent = siblingIndex
	grandChild.parent = parentIndex

	// The sibling adopted a new child, so its box and height are stale. The
	// parent is refreshed by the caller, since its box cannot have changed.
	t.refreshNode(siblingIndex)
}

const (
	nullBVHIndex = int32(-1)

	// freeBVHNodeHeight marks the height of a node that is not currently part
	// of the tree and is waiting on the free list.
	freeBVHNodeHeight = int32(-1)

	// bvhRelativeMargin is the fraction of the radius of an item that is used
	// as a fattening margin, whenever it exceeds the margin that was configured
	// through [BVHTreeSettings.AABBMargin].
	bvhRelativeMargin = 0.05

	// bvhStackCapacity is the capacity of the buffer that query traversal uses
	// as a stack. Rotations keep the height of the tree at roughly 1.6*log2(N),
	// which leaves plenty of room for any realistic item count. A tree that
	// somehow grew deeper would still be traversed correctly, at the cost of
	// moving the stack to the heap.
	bvhStackCapacity = 64
)

// bvhNode is a node of a [BVHTree].
//
// The struct is deliberately kept at exactly 64 bytes, so that a node occupies
// a single cache line. To achieve that, leaf nodes reuse the childRight field
// in order to store the index of their item. Use [bvhNode.isLeaf] to tell the
// two cases apart.
type bvhNode struct {
	box        treeAABB // grown box for leaf nodes, merged box for the rest
	parent     int32    // nullBVHIndex for the root node
	childLeft  int32    // nullBVHIndex for leaf nodes
	childRight int32    // the item index for leaf nodes
	height     int32    // zero for leaf nodes
}

func (n *bvhNode) isLeaf() bool {
	return n.childLeft == nullBVHIndex
}

type bvhItem[T any] struct {
	box   treeAABB // the exact box of the item, not the grown one
	node  int32    // nullBVHIndex once the item has been removed
	value T
}

// bvhRay is a precomputed form of a [Segment] that allows the slab test to use
// multiplications instead of divisions.
type bvhRay struct {
	originX   float64
	originY   float64
	originZ   float64
	invDeltaX float64
	invDeltaY float64
	invDeltaZ float64
}

func newBVHRay(segment *Segment) bvhRay {
	return bvhRay{
		originX:   segment.a.X,
		originY:   segment.a.Y,
		originZ:   segment.a.Z,
		invDeltaX: bvhInverse(segment.b.X - segment.a.X),
		invDeltaY: bvhInverse(segment.b.Y - segment.a.Y),
		invDeltaZ: bvhInverse(segment.b.Z - segment.a.Z),
	}
}

// bvhInverse returns the reciprocal of the specified value, mapping a zero
// value to a very large but finite result.
//
// Using a finite value instead of an infinity yields the same accept and
// reject decisions for a degenerate axis, without the risk of producing a NaN
// through a multiplication with a zero distance.
func bvhInverse(delta float64) float64 {
	if delta == 0.0 {
		return math.MaxFloat64
	}
	return 1.0 / delta
}
