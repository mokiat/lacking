package query2d

import (
	"log/slog"
	"math"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
)

// InvalidQuadtreeItemID is an identifier that can be used by user
// code to mark an identifier as invalid. Such an identifier will
// never be returned by the library but must also never be passed to the
// library.
const InvalidQuadtreeItemID = QuadtreeItemID(0xFFFFFFFF)

// QuadtreeItemID is an identifier used to control the placement of an item
// into a tree.
type QuadtreeItemID uint32

// QuadtreeSettings contains the settings for a Quadtree.
type QuadtreeSettings struct {

	// Size specifies the dimension (side to side) of the tree node.
	//
	// If not specified, a default size of 4096 is used.
	//
	// Inserting an item outside these bounds has undefined behavior.
	Size opt.T[float32]

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

// Quadtree is a spatial structure that uses a loose quadtree implementation
// with shrinking bounding box to enable the fast searching of items.
type Quadtree[T any] struct {
	nodes           []quadtreeNode
	items           []quadtreeItem[T]
	freeNodeIndices *ds.Stack[int32]
	freeItemIDs     *ds.Stack[QuadtreeItemID]
	idMappings      []int32
	maxDepth        uint32

	nodeCountAccepted uint32
	nodeCountRejected uint32
	itemCountAccepted uint32
	itemCountRejected uint32

	isDirty bool
}

// NewQuadtree creates a new Quadtree using the provided settings.
func NewQuadtree[T any](settings QuadtreeSettings) *Quadtree[T] {
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

	nodes := make([]quadtreeNode, 0, initialNodeCapacity)
	nodes = append(nodes, quadtreeNode{
		parent:    nullQuadtreeIndex,
		children:  emptyQuadtreeNodeChildren,
		itemStart: 0,
		itemEnd:   0,
		looseArea: quadtreeQuad{
			x: 0.0,
			y: 0.0,
			r: float32(size), // using size here since a loose area has twice the size
		},
		box: emptyQuadtreeAABB(),
	})

	return &Quadtree[T]{
		nodes:           nodes,
		items:           make([]quadtreeItem[T], 0, initialItemCapacity),
		freeNodeIndices: ds.EmptyStack[int32](),
		freeItemIDs:     ds.EmptyStack[QuadtreeItemID](),
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
func (t *Quadtree[T]) Stats() TreeStats {
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
func (t *Quadtree[T]) VisitStats() TreeVisitStats {
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
func (t *Quadtree[T]) Insert(area Area, value T) QuadtreeItemID {
	node := t.pickNodeForItem(area)
	box := newQuadtreeAABBFromArea(area)
	t.markNodeDirty(node)

	if t.freeItemIDs.IsEmpty() {
		if len(t.items) == cap(t.items) {
			logger.Warn("Growing item capacity for tree.", slog.Int("current", len(t.items)))
		}
		id := QuadtreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		t.items = append(t.items, quadtreeItem[T]{
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
func (t *Quadtree[T]) Update(id QuadtreeItemID, area Area) {
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == nullQuadtreeIndex {
		panic("cannot update removed item")
	}
	item.box = newQuadtreeAABBFromArea(area)
	t.markNodeDirty(item.node) // previous node
	item.node = t.pickNodeForItem(area)
	t.markNodeDirty(item.node) // new node
}

// Remove removes the item with the specified id from this tree.
func (t *Quadtree[T]) Remove(id QuadtreeItemID) {
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == nullQuadtreeIndex {
		panic("cannot remove item twice")
	}
	t.markNodeDirty(item.node)
	item.node = nullQuadtreeIndex
	t.freeItemIDs.Push(id)
}

// QuerySegment finds all items that intersect the specified segment. Each
// found item is passed to the specified yield function. The order in which
// items are passed is undefined and might change between invocations.
func (t *Quadtree[T]) QuerySegment(segment Segment, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInSegment(0, &segment, yield)
}

// QueryAABB finds all items that are inside or intersect the specified
// axis-aligned bounding box. Each found item is passed to the specified yield
// function. The order in which items are passed is undefined and might change
// between invocations.
func (t *Quadtree[T]) QueryAABB(aabb AABB, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInAABB(0, &aabb, yield)
}

// GC runs cleanup and optimization logic. You should call this at least once
// per frame.
func (t *Quadtree[T]) GC() {
	t.refresh()
}

func (t *Quadtree[T]) resetVisitStats() {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
}

func (t *Quadtree[T]) activeNodeCount() uint32 {
	return uint32(len(t.nodes) - t.freeNodeIndices.Size())
}

func (t *Quadtree[T]) activeItemCount() uint32 {
	return uint32(len(t.items) - t.freeItemIDs.Size())
}

func (t *Quadtree[T]) markNodeDirty(nodeIndex int32) {
	t.isDirty = true
	node := &t.nodes[nodeIndex]
	node.isDirty = true
}

func (t *Quadtree[T]) itemsAtDepth(nodeIndex int32, currentDepth, depth uint32) uint32 {
	if nodeIndex == nullQuadtreeIndex {
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

func (t *Quadtree[T]) pickNodeForItem(area Area) int32 {
	bestNodeIndex := nullQuadtreeIndex
	currentNodeIndex := int32(0)
	var depth uint32
	for currentNodeIndex != nullQuadtreeIndex {
		bestNodeIndex = currentNodeIndex
		depth++
		if depth >= t.maxDepth {
			break
		}
		currentNodeIndex = t.pickChildNode(currentNodeIndex, area)
	}
	return bestNodeIndex
}

func (t *Quadtree[T]) pickChildNode(parentNodeIndex int32, area Area) int32 {
	parentNode := &t.nodes[parentNodeIndex]
	parentLooseArea := parentNode.looseArea

	// Make sure that it can fit inside a child. The requirement is that
	// the radius must be smaller than the loose margin of the child.
	childLooseRadius := parentLooseArea.r / 2.0
	if area.r > (childLooseRadius / 2.0) { // div by 2 to convert to margin
		return nullQuadtreeIndex
	}

	// It has to be inside one of the four children.
	var (
		childIndex = 0
		childX     = parentLooseArea.x
		childY     = parentLooseArea.y
	)
	childOffset := parentLooseArea.r / 4.0
	if area.x < parentLooseArea.x {
		childX -= childOffset
	} else {
		childIndex += 1
		childX += childOffset
	}
	if area.y < parentLooseArea.y {
		childY -= childOffset
	} else {
		childIndex += 2
		childY += childOffset
	}

	if parentNode.children[childIndex] != nullQuadtreeIndex {
		return parentNode.children[childIndex]
	}

	childLooseArea := quadtreeQuad{
		x: childX,
		y: childY,
		r: childLooseRadius,
	}
	if t.freeNodeIndices.IsEmpty() {
		if len(t.nodes) == cap(t.nodes) {
			logger.Warn("Growing node capacity for tree.", slog.Int("current", len(t.nodes)))
		}
		childNodeIndex := int32(len(t.nodes)) // predict next node index
		parentNode.children[childIndex] = childNodeIndex
		// Do NOT use "parentNode" after this append as the ref might be towards
		// an old slice!
		t.nodes = append(t.nodes, quadtreeNode{
			parent:    parentNodeIndex,
			children:  emptyQuadtreeNodeChildren,
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
		childNode.children = emptyQuadtreeNodeChildren
		childNode.looseArea = childLooseArea
		childNode.itemStart = 0
		childNode.itemEnd = 0
		return childNodeIndex
	}
}

func (t *Quadtree[T]) refresh() {
	if t.isDirty {
		t.groupItems()
		t.updateIDMappings()
		t.gcNodes()
		t.updateAABB(0)
		t.isDirty = false
	}
}

func (t *Quadtree[T]) groupItems() {
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart = 0
		node.itemEnd = 0
		node.sortEnd = 0
	}
	for i := range t.items {
		item := &t.items[i]
		if item.node != nullQuadtreeIndex {
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
		if item.node == nullQuadtreeIndex {
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

func (t *Quadtree[T]) swapItems(i, j uint32) {
	if i != j {
		t.items[i], t.items[j] = t.items[j], t.items[i]
	}
}

func (t *Quadtree[T]) updateIDMappings() {
	for i, item := range t.items {
		t.idMappings[item.id] = int32(i)
	}
}

func (t *Quadtree[T]) gcNodes() {
	for i := range t.nodes {
		t.gcNode(int32(i))
	}
}

func (t *Quadtree[T]) gcNode(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	if node.parent == nullQuadtreeIndex {
		return // already deleted or root
	}
	if !node.isEmpty() {
		return // can't gc node
	}
	parentNodeIndex := node.parent
	parentNode := &t.nodes[parentNodeIndex]
	for i, childNodeIndex := range parentNode.children {
		if childNodeIndex == nodeIndex {
			parentNode.children[i] = nullQuadtreeIndex
			break
		}
	}
	node.parent = nullQuadtreeIndex
	t.freeNodeIndices.Push(nodeIndex)
	t.gcNode(parentNodeIndex)
}

func (t *Quadtree[T]) updateAABB(nodeIndex int32) bool {
	node := &t.nodes[nodeIndex]

	var wereChildrenDirty bool
	for _, childIndex := range node.children {
		if childIndex != nullQuadtreeIndex {
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

	result := emptyQuadtreeAABB()
	for _, childIndex := range node.children {
		if childIndex != nullQuadtreeIndex {
			child := &t.nodes[childIndex]
			result = mergeQuadtreeAABBs(result, child.box)
		}
	}
	for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
		item := &t.items[itemIndex]
		result = mergeQuadtreeAABBs(result, item.box)
	}
	node.box = result
	node.isDirty = false

	return true
}

func (t *Quadtree[T]) visitNodeInSegment(nodeIndex int32, querySegment *Segment, yield VisitorFunc[T]) bool {
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
			if childNodeIndex != nullQuadtreeIndex {
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

func (t *Quadtree[T]) visitNodeInAABB(nodeIndex int32, queryAABB *AABB, yield VisitorFunc[T]) bool {
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
			if childNodeIndex != nullQuadtreeIndex {
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

const nullQuadtreeIndex = int32(-1)

var emptyQuadtreeNodeChildren = [4]int32{
	nullQuadtreeIndex, nullQuadtreeIndex,
	nullQuadtreeIndex, nullQuadtreeIndex,
}

type quadtreeNode struct {
	parent    int32
	children  [4]int32
	looseArea quadtreeQuad
	box       quadtreeAABB
	itemStart uint32
	itemEnd   uint32
	sortEnd   uint32
	isDirty   bool
}

func (n *quadtreeNode) isEmpty() bool {
	return (n.children == emptyQuadtreeNodeChildren) && (n.itemStart >= n.itemEnd)
}

type quadtreeItem[T any] struct {
	id    QuadtreeItemID
	node  int32
	box   quadtreeAABB
	value T
}

type quadtreeQuad struct {
	x float32
	y float32
	r float32
}

type quadtreeAABB struct {
	minX float32
	maxX float32
	minY float32
	maxY float32
}

func emptyQuadtreeAABB() quadtreeAABB {
	return quadtreeAABB{
		minX: math.MaxFloat32,
		minY: math.MaxFloat32,
		maxX: -math.MaxFloat32,
		maxY: -math.MaxFloat32,
	}
}

func newQuadtreeAABBFromArea(area Area) quadtreeAABB {
	return quadtreeAABB{
		minX: area.x - area.r,
		minY: area.y - area.r,
		maxX: area.x + area.r,
		maxY: area.y + area.r,
	}
}

func mergeQuadtreeAABBs(first, second quadtreeAABB) quadtreeAABB {
	return quadtreeAABB{
		minX: min(first.minX, second.minX),
		minY: min(first.minY, second.minY),
		maxX: max(first.maxX, second.maxX),
		maxY: max(first.maxY, second.maxY),
	}
}

func (aabb *quadtreeAABB) isEmpty() bool {
	return (aabb.minX > aabb.maxX) || (aabb.minY > aabb.maxY)
}

func (aabb *quadtreeAABB) intersectsSegment(segment *Segment) bool {
	if aabb.isEmpty() {
		return false
	}

	delta := sprec.Vec2Diff(segment.b, segment.a)

	var tCloseX, tFarX float32
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

	var tCloseY, tFarY float32
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

	tClose := max(tCloseX, tCloseY)
	tFar := min(tFarX, tFarY)

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}

func (aabb *quadtreeAABB) intersectsAABB(other *AABB) bool {
	if aabb.isEmpty() {
		return false
	}
	return (aabb.minX <= other.maxX) &&
		(aabb.minY <= other.maxY) &&
		(aabb.maxX >= other.minX) &&
		(aabb.maxY >= other.minY)
}
