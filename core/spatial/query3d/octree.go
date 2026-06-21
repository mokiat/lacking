package query3d

import (
	"math"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// OctreeSettings contains the settings for a Octree.
type OctreeSettings struct {

	// Size specifies the dimension (side to side) of the tree node.
	//
	// If not specified, a default size of 4096 is used.
	//
	// Inserting an item outside these bounds has undefined behavior.
	Size opt.T[float64]

	// MaxDepth controls the maximum depth that the tree can reach.
	//
	// If not specified, a default max depth of 8 is used.
	MaxDepth opt.T[uint32]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. Usually one would find that number empirically.
	// This allows the tree to preallocate memory and avoid dynamic allocations.
	//
	// By default the initial capacity is 4096.
	InitialNodeCapacity opt.T[uint32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the tree. This allows the tree to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[uint32]
}

// Octree is a spatial structure that uses a loose octree implementation
// with shrinking bounding box to enable the fast searching of items.
type Octree[T any] struct {
	nodes           []octreeNode
	items           []octreeItem[T]
	freeNodeIndices *ds.Stack[int32]
	freeItemIDs     *ds.Stack[TreeItemID]
	idMappings      []int32
	maxDepth        uint32

	nodeCountAccepted uint32
	nodeCountRejected uint32
	itemCountAccepted uint32
	itemCountRejected uint32

	isDirty bool
}

// NewOctree creates a new Octree using the provided settings.
func NewOctree[T any](settings OctreeSettings) *Octree[T] {
	size := settings.Size.ValueOrDefault(4096.0)
	if size < 1.0 {
		panic("size cannot be smaller than 1.0")
	}
	maxDepth := settings.MaxDepth.ValueOrDefault(8)
	if maxDepth == 0 {
		panic("max depth cannot be zero")
	}
	initialNodeCapacity := settings.InitialNodeCapacity.ValueOrDefault(4096)
	initialItemCapacity := settings.InitialItemCapacity.ValueOrDefault(1024)

	nodes := make([]octreeNode, 0, initialNodeCapacity)
	nodes = append(nodes, octreeNode{
		parent:    nullOctreeIndex,
		children:  emptyOctreeNodeChildren,
		itemStart: 0,
		itemEnd:   0,
		looseArea: octreeCube{
			x: 0.0,
			y: 0.0,
			z: 0.0,
			r: size, // using size here since a loose area has twice the size
		},
		box: emptyOctreeAABB(),
	})

	return &Octree[T]{
		nodes:           nodes,
		items:           make([]octreeItem[T], 0, initialItemCapacity),
		freeNodeIndices: ds.EmptyStack[int32](),
		freeItemIDs:     ds.EmptyStack[TreeItemID](),
		idMappings:      make([]int32, 0, initialItemCapacity),
		maxDepth:        maxDepth,

		nodeCountAccepted: 0,
		nodeCountRejected: 0,
		itemCountAccepted: 0,
		itemCountRejected: 0,

		isDirty: false,
	}
}

// Stats returns statistics on the current state of this tree.
func (t *Octree[T]) Stats() TreeStats {
	t.refresh() // this is necessary
	itemCountPerDepth := make([]uint32, t.maxDepth)
	for i := range t.maxDepth {
		itemCountPerDepth[i] = t.itemsAtDepth(0, 1, i+1)
	}
	return TreeStats{
		NodeCount:         t.activeNodeCount(),
		ItemCount:         t.activeItemCount(),
		ItemCountPerDepth: itemCountPerDepth,
	}
}

// VisitStats returns statistics information on the last executed search in
// this tree.
func (t *Octree[T]) VisitStats() TreeVisitStats {
	return TreeVisitStats{
		NodeCountVisited:  t.nodeCountAccepted + t.nodeCountRejected,
		NodeCountAccepted: t.nodeCountAccepted,
		NodeCountRejected: t.nodeCountRejected,
		ItemCountVisited:  t.itemCountAccepted + t.itemCountRejected,
		ItemCountAccepted: t.itemCountAccepted,
		ItemCountRejected: t.itemCountRejected,
	}
}

// Insert adds an item, which occupies the specified area, to this
// tree.
func (t *Octree[T]) Insert(area Area, value T) TreeItemID {
	node := t.pickNodeForItem(area)
	box := newOctreeAABBFromArea(area)
	t.markNodeDirty(node)

	if t.freeItemIDs.IsEmpty() {
		id := TreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		t.items = append(t.items, octreeItem[T]{
			id:    id,
			node:  node,
			box:   box,
			value: value,
		})
		return id
	} else {
		id := t.freeItemIDs.Pop()
		itemIndex := t.idMappings[id]
		item := &t.items[itemIndex]
		item.box = box
		item.value = value
		item.node = node
		return item.id
	}
}

// Update repositions the item with the specified id to the new area.
func (t *Octree[T]) Update(id TreeItemID, area Area) {
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == nullOctreeIndex {
		panic("cannot update removed item")
	}
	item.box = newOctreeAABBFromArea(area)
	t.markNodeDirty(item.node) // previous node
	item.node = t.pickNodeForItem(area)
	t.markNodeDirty(item.node) // new node
}

// Remove removes the item with the specified id from this tree.
func (t *Octree[T]) Remove(id TreeItemID) {
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == nullOctreeIndex {
		panic("cannot remove item twice")
	}
	t.markNodeDirty(item.node)
	item.node = nullOctreeIndex
	t.freeItemIDs.Push(id)
}

// QuerySegment finds all items that intersect the specified segment. Each
// found item is passed to the specified yield function. The order in which
// items are passed is undefined and might change between invocations.
func (t *Octree[T]) QuerySegment(segment Segment, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInSegment(0, &segment, yield)
}

// QueryAABB finds all items that are inside or intersect the specified
// axis-aligned bounding box. Each found item is passed to the specified yield
// function. The order in which items are passed is undefined and might change
// between invocations.
func (t *Octree[T]) QueryAABB(aabb AABB, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInAABB(0, &aabb, yield)
}

// GC runs cleanup and optimization logic. You should call this at least once
// per frame.
func (t *Octree[T]) GC() {
	t.refresh()
}

func (t *Octree[T]) resetVisitStats() {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
}

func (t *Octree[T]) activeNodeCount() uint32 {
	return uint32(len(t.nodes) - t.freeNodeIndices.Size())
}

func (t *Octree[T]) activeItemCount() uint32 {
	return uint32(len(t.items) - t.freeItemIDs.Size())
}

func (t *Octree[T]) markNodeDirty(nodeIndex int32) {
	t.isDirty = true
	node := &t.nodes[nodeIndex]
	node.isDirty = true
}

func (t *Octree[T]) itemsAtDepth(nodeIndex int32, currentDepth, depth uint32) uint32 {
	if nodeIndex == nullOctreeIndex {
		return 0
	}
	node := &t.nodes[nodeIndex]
	if currentDepth == depth {
		return node.itemEnd - node.itemStart
	}
	var result uint32
	for _, childNodeIndex := range node.children {
		result += t.itemsAtDepth(childNodeIndex, currentDepth+1, depth)
	}
	return result
}

func (t *Octree[T]) pickNodeForItem(area Area) int32 {
	bestNodeIndex := nullOctreeIndex
	currentNodeIndex := int32(0)
	var depth uint32
	for currentNodeIndex != nullOctreeIndex {
		bestNodeIndex = currentNodeIndex
		depth++
		if depth >= t.maxDepth {
			break
		}
		currentNodeIndex = t.pickChildNode(currentNodeIndex, area)
	}
	return bestNodeIndex
}

func (t *Octree[T]) pickChildNode(parentNodeIndex int32, area Area) int32 {
	parentNode := &t.nodes[parentNodeIndex]
	parentLooseArea := parentNode.looseArea

	// Make sure that it can fit inside a child. The requirement is that
	// the radius must be smaller than the loose margin of the child.
	childLooseRadius := parentLooseArea.r / 2.0
	if area.r > (childLooseRadius / 2.0) { // div by 2 to convert to margin
		return nullOctreeIndex
	}

	// It has to be inside one of the four children.
	var (
		childIndex = 0
		childX     = parentLooseArea.x
		childY     = parentLooseArea.y
		childZ     = parentLooseArea.z
	)
	childOffset := parentLooseArea.r / 4.0
	if area.x < parentLooseArea.x {
		childX -= childOffset
	} else {
		childIndex += 1
		childX += childOffset
	}
	if area.z < parentLooseArea.z {
		childZ -= childOffset
	} else {
		childIndex += 2
		childZ += childOffset
	}
	if area.y < parentLooseArea.y {
		childY -= childOffset
	} else {
		childIndex += 4
		childY += childOffset
	}

	if parentNode.children[childIndex] != nullOctreeIndex {
		return parentNode.children[childIndex]
	}

	childLooseArea := octreeCube{
		x: childX,
		y: childY,
		z: childZ,
		r: childLooseRadius,
	}
	if t.freeNodeIndices.IsEmpty() {
		childNodeIndex := int32(len(t.nodes)) // predict next node index
		parentNode.children[childIndex] = childNodeIndex
		// Do NOT use "parentNode" after this append as the ref might be towards
		// an old slice!
		t.nodes = append(t.nodes, octreeNode{
			parent:    parentNodeIndex,
			children:  emptyOctreeNodeChildren,
			looseArea: childLooseArea,
			itemStart: 0,
			itemEnd:   0,
		})
		return childNodeIndex
	} else {
		childNodeIndex := t.freeNodeIndices.Pop()
		parentNode.children[childIndex] = childNodeIndex
		childNode := &t.nodes[childNodeIndex]
		childNode.parent = parentNodeIndex
		childNode.children = emptyOctreeNodeChildren
		childNode.looseArea = childLooseArea
		childNode.itemStart = 0
		childNode.itemEnd = 0
		return childNodeIndex
	}
}

func (t *Octree[T]) refresh() {
	if t.isDirty {
		t.groupItems()
		t.updateIDMappings()
		t.gcNodes()
		t.updateAABB(0)
		t.isDirty = false
	}
}

func (t *Octree[T]) groupItems() {
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart = 0
		node.itemEnd = 0
		node.sortEnd = 0
	}
	for i := range t.items {
		item := &t.items[i]
		if item.node != nullOctreeIndex {
			node := &t.nodes[item.node]
			node.itemEnd++ // use as counter for now
		}
	}
	offset := uint32(0)
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart += offset
		node.itemEnd += offset
		node.sortEnd = node.itemStart
		offset = node.itemEnd
	}
	countActiveItems := uint32(offset)
	for i := uint32(0); i < countActiveItems; {
		item := &t.items[i]
		if item.node == nullOctreeIndex {
			t.swapItems(i, offset)
			offset++
			continue
		}
		node := &t.nodes[item.node]
		if i >= node.itemStart && i < node.sortEnd {
			i++ // item is in the right place
			continue
		}
		t.swapItems(i, node.sortEnd)
		node.sortEnd++
	}
}

func (t *Octree[T]) swapItems(i, j uint32) {
	if i != j {
		t.items[i], t.items[j] = t.items[j], t.items[i]
	}
}

func (t *Octree[T]) updateIDMappings() {
	for i, item := range t.items {
		t.idMappings[item.id] = int32(i)
	}
}

func (t *Octree[T]) gcNodes() {
	for i := range t.nodes {
		t.gcNode(int32(i))
	}
}

func (t *Octree[T]) gcNode(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	if node.parent == nullOctreeIndex {
		return // already deleted or root
	}
	if !node.isEmpty() {
		return // can't gc node
	}
	parentNodeIndex := node.parent
	parentNode := &t.nodes[parentNodeIndex]
	for i, childNodeIndex := range parentNode.children {
		if childNodeIndex == nodeIndex {
			parentNode.children[i] = nullOctreeIndex
			break
		}
	}
	node.parent = nullOctreeIndex
	t.freeNodeIndices.Push(nodeIndex)
	t.gcNode(parentNodeIndex)
}

func (t *Octree[T]) updateAABB(nodeIndex int32) bool {
	node := &t.nodes[nodeIndex]

	var wereChildrenDirty bool
	for _, childIndex := range node.children {
		if childIndex != nullOctreeIndex {
			if t.updateAABB(childIndex) {
				wereChildrenDirty = true
			}
		}
	}

	if !node.isDirty && !wereChildrenDirty {
		return false
	}

	// One potential optimization is to split the box cache into two parts:
	// - one for the items boxes
	// - one for overall (current)
	// Depending on node.isDirty the overall box can be recomputed from the
	// cached items boxes. This would avoid recomputing the items boxes every
	// time.

	result := emptyOctreeAABB()
	for _, childIndex := range node.children {
		if childIndex != nullOctreeIndex {
			child := &t.nodes[childIndex]
			result = mergeOctreeAABBs(result, child.box)
		}
	}
	for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
		item := &t.items[itemIndex]
		result = mergeOctreeAABBs(result, item.box)
	}
	node.box = result
	node.isDirty = false

	return true
}

func (t *Octree[T]) visitNodeInSegment(nodeIndex int32, querySegment *Segment, yield VisitorFunc[T]) bool {
	node := &t.nodes[nodeIndex]
	if node.box.intersectsSegment(querySegment) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if item.box.intersectsSegment(querySegment) {
				t.itemCountAccepted++
				if !yield(item.value) {
					return false
				}
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			if childNodeIndex != nullOctreeIndex {
				if !t.visitNodeInSegment(childNodeIndex, querySegment, yield) {
					return false
				}
			}
		}
	} else {
		t.nodeCountRejected++
	}
	return true
}

func (t *Octree[T]) visitNodeInAABB(nodeIndex int32, queryAABB *AABB, yield VisitorFunc[T]) bool {
	node := &t.nodes[nodeIndex]
	if node.box.intersectsAABB(queryAABB) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if item.box.intersectsAABB(queryAABB) {
				t.itemCountAccepted++
				if !yield(item.value) {
					return false
				}
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			if childNodeIndex != nullOctreeIndex {
				if !t.visitNodeInAABB(childNodeIndex, queryAABB, yield) {
					return false
				}
			}
		}
	} else {
		t.nodeCountRejected++
	}
	return true
}

const nullOctreeIndex = int32(-1)

var emptyOctreeNodeChildren = [8]int32{
	nullOctreeIndex, nullOctreeIndex,
	nullOctreeIndex, nullOctreeIndex,

	nullOctreeIndex, nullOctreeIndex,
	nullOctreeIndex, nullOctreeIndex,
}

type octreeNode struct {
	parent    int32
	children  [8]int32
	looseArea octreeCube
	box       octreeAABB
	itemStart uint32
	itemEnd   uint32
	sortEnd   uint32
	isDirty   bool
}

func (n *octreeNode) isEmpty() bool {
	return (n.children == emptyOctreeNodeChildren) && (n.itemStart >= n.itemEnd)
}

type octreeItem[T any] struct {
	id    TreeItemID
	node  int32
	box   octreeAABB
	value T
}

type octreeCube struct {
	x float64
	y float64
	z float64
	r float64
}

type octreeAABB struct {
	minX float64
	minY float64
	minZ float64
	maxX float64
	maxY float64
	maxZ float64
}

func emptyOctreeAABB() octreeAABB {
	return octreeAABB{
		minX: math.MaxFloat64,
		minY: math.MaxFloat64,
		minZ: math.MaxFloat64,
		maxX: -math.MaxFloat64,
		maxY: -math.MaxFloat64,
		maxZ: -math.MaxFloat64,
	}
}

func newOctreeAABBFromArea(area Area) octreeAABB {
	return octreeAABB{
		minX: area.x - area.r,
		minY: area.y - area.r,
		minZ: area.z - area.r,
		maxX: area.x + area.r,
		maxY: area.y + area.r,
		maxZ: area.z + area.r,
	}
}

func mergeOctreeAABBs(first, second octreeAABB) octreeAABB {
	return octreeAABB{
		minX: min(first.minX, second.minX),
		minY: min(first.minY, second.minY),
		minZ: min(first.minZ, second.minZ),
		maxX: max(first.maxX, second.maxX),
		maxY: max(first.maxY, second.maxY),
		maxZ: max(first.maxZ, second.maxZ),
	}
}

func (aabb *octreeAABB) isEmpty() bool {
	return (aabb.minX > aabb.maxX) || (aabb.minY > aabb.maxY) || (aabb.minZ > aabb.maxZ)
}

func (aabb *octreeAABB) intersectsSegment(segment *Segment) bool {
	if aabb.isEmpty() {
		return false
	}

	delta := dprec.Vec3Diff(segment.b, segment.a)

	var tCloseX, tFarX float64
	if delta.X == 0.0 {
		if (segment.a.X < aabb.minX) || (segment.a.X > aabb.maxX) {
			return false // // both points are outside the box on the left or right
		}
		tCloseX = 0.0
		tFarX = 1.0
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
		tCloseY = 0.0
		tFarY = 1.0
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
		tCloseZ = 0.0
		tFarZ = 1.0
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

func (aabb *octreeAABB) intersectsAABB(other *AABB) bool {
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
