package query2d

import (
	"math"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// QuadtreeSettings contains the settings for a [Quadtree].
type QuadtreeSettings struct {

	// Size specifies the dimension (side to side) of the tree node.
	//
	// If not specified, a default size of 4096 is used.
	//
	// Items that lie outside these bounds are still found by queries, but they
	// settle in nodes close to the root and therefore degrade query
	// performance.
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

// Quadtree is a spatial structure that uses a loose quadtree implementation
// with shrinking bounding box to enable the fast searching of items.
type Quadtree[T any] struct {
	nodes           []quadtreeNode
	items           []quadtreeItem[T]
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

// NewQuadtree creates a new [Quadtree] using the provided settings.
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
		parent:      nullQuadtreeIndex,
		children:    emptyQuadtreeNodeChildren,
		itemCount:   0,
		itemOffset:  0,
		placeOffset: 0,
		looseArea: quadtreeQuad{
			x:        0.0,
			y:        0.0,
			halfSize: size, // using size here since a loose area has twice the size
		},
		tightArea: emptyQuadtreeAABB(),
	})

	return &Quadtree[T]{
		nodes:           nodes,
		items:           make([]quadtreeItem[T], 0, initialItemCapacity),
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

// Insert adds an item, which occupies the specified axis-aligned bounding
// box, to this tree.
//
// The item is placed in the deepest node that can fully contain the box, so
// the smaller and the better centered the box is, the less work queries have
// to do. The box must not be empty (as per [shape2d.AABB.IsEmpty]), otherwise
// this function panics.
func (t *Quadtree[T]) Insert(aabb shape2d.AABB, value T) TreeItemID {
	if aabb.IsEmpty() {
		panic("cannot insert item with empty area")
	}

	tightArea := newQuadtreeAABBFromAABB(aabb)
	nodeIndex := t.pickNodeForItem(tightArea)
	t.increaseNodeItems(nodeIndex)

	if t.freeItemIDs.IsEmpty() {
		id := TreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		t.items = append(t.items, quadtreeItem[T]{
			id:        id,
			node:      nodeIndex,
			tightArea: tightArea,
			value:     value,
		})
		return id
	} else {
		id := t.freeItemIDs.Pop()
		itemIndex := t.idMappings[id]
		item := &t.items[itemIndex]
		item.tightArea = tightArea
		item.value = value
		item.node = nodeIndex
		return item.id
	}
}

// Update repositions and resizes the item with the specified id to the new
// axis-aligned bounding box.
//
// As with [Quadtree.Insert], the box must not be empty, otherwise this
// function panics. Updating an item that has already been removed panics as
// well.
func (t *Quadtree[T]) Update(id TreeItemID, aabb shape2d.AABB) {
	if aabb.IsEmpty() {
		panic("cannot update item to empty area")
	}

	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == nullQuadtreeIndex {
		panic("cannot update removed item")
	}
	tightArea := newQuadtreeAABBFromAABB(aabb)
	item.tightArea = tightArea
	oldNodeIndex := item.node
	t.decreaseNodeItems(item.node) // previous node
	item.node = t.pickNodeForItem(tightArea)
	t.increaseNodeItems(item.node) // new node
	t.gcNode(oldNodeIndex)
}

// Remove removes the item with the specified id from this tree.
func (t *Quadtree[T]) Remove(id TreeItemID) {
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == nullQuadtreeIndex {
		panic("cannot remove item twice")
	}
	oldNodeIndex := item.node
	t.decreaseNodeItems(item.node)
	item.node = nullQuadtreeIndex
	t.freeItemIDs.Push(id)
	t.gcNode(oldNodeIndex)
}

// QuerySegment finds all items that intersect the specified segment. Each
// found item is passed to the specified yield function. The order in which
// items are passed is undefined and might change between invocations.
func (t *Quadtree[T]) QuerySegment(segment shape2d.Segment, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInSegment(0, &segment, yield)
}

// QueryAABB finds all items that are inside or intersect the specified
// axis-aligned bounding box. Each found item is passed to the specified yield
// function. The order in which items are passed is undefined and might change
// between invocations.
func (t *Quadtree[T]) QueryAABB(aabb shape2d.AABB, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInAABB(0, &aabb, yield)
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

func (t *Quadtree[T]) increaseNodeItems(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	node.itemCount++
	node.isDirty = true
	t.isDirty = true
}

func (t *Quadtree[T]) decreaseNodeItems(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	node.itemCount--
	node.isDirty = true
	t.isDirty = true
}

func (t *Quadtree[T]) itemsAtDepth(nodeIndex int32, currentDepth, depth uint32) uint32 {
	if nodeIndex == nullQuadtreeIndex {
		return 0
	}
	node := &t.nodes[nodeIndex]
	if currentDepth == depth {
		return node.itemCount
	}
	var result uint32
	for _, childNodeIndex := range node.children {
		result += t.itemsAtDepth(childNodeIndex, currentDepth+1, depth)
	}
	return result
}

// pickNodeForItem returns the deepest node whose loose area still fully
// contains the specified area.
func (t *Quadtree[T]) pickNodeForItem(area quadtreeAABB) int32 {
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

// pickChildNode returns the child of the specified node whose loose area fully
// contains the specified area, allocating that child if it does not exist yet.
// It returns nullQuadtreeIndex if the area does not fit in any child.
func (t *Quadtree[T]) pickChildNode(parentNodeIndex int32, area quadtreeAABB) int32 {
	parentNode := &t.nodes[parentNodeIndex]
	parentLooseArea := parentNode.looseArea

	// The candidate child is the one whose own (tight) quadrant holds the
	// center of the area.
	const half = 1.0 / 2.0
	var (
		childIndex = 0
		childX     = parentLooseArea.x
		childY     = parentLooseArea.y
	)
	childOffset := parentLooseArea.halfSize / 4.0
	if (area.minX+area.maxX)*half < parentLooseArea.x {
		childX -= childOffset
	} else {
		childIndex += 1
		childX += childOffset
	}
	if (area.minY+area.maxY)*half < parentLooseArea.y {
		childY -= childOffset
	} else {
		childIndex += 2
		childY += childOffset
	}

	// The area has to fit within the loose area of that child. Each axis is
	// checked against its own extent, so an elongated or a well-centered item
	// is no longer held back by its largest dimension and can descend deeper
	// than a bounding-square test would allow.
	childLooseHalfSize := parentLooseArea.halfSize * half
	if (area.minX < childX-childLooseHalfSize) || (area.maxX > childX+childLooseHalfSize) {
		return nullQuadtreeIndex
	}
	if (area.minY < childY-childLooseHalfSize) || (area.maxY > childY+childLooseHalfSize) {
		return nullQuadtreeIndex
	}

	if parentNode.children[childIndex] != nullQuadtreeIndex {
		return parentNode.children[childIndex]
	}

	childLooseArea := quadtreeQuad{
		x:        childX,
		y:        childY,
		halfSize: childLooseHalfSize,
	}
	if t.freeNodeIndices.IsEmpty() {
		childNodeIndex := int32(len(t.nodes)) // predict next node index
		parentNode.children[childIndex] = childNodeIndex
		// Do NOT use "parentNode" after this append as the ref might be towards
		// an old slice!
		t.nodes = append(t.nodes, quadtreeNode{
			parent:      parentNodeIndex,
			children:    emptyQuadtreeNodeChildren,
			looseArea:   childLooseArea,
			itemCount:   0,
			itemOffset:  0,
			placeOffset: 0,
		})
		return childNodeIndex
	} else {
		childNodeIndex := t.freeNodeIndices.Pop()
		parentNode.children[childIndex] = childNodeIndex
		childNode := &t.nodes[childNodeIndex]
		childNode.parent = parentNodeIndex
		childNode.children = emptyQuadtreeNodeChildren
		childNode.looseArea = childLooseArea
		childNode.itemCount = 0
		childNode.itemOffset = 0
		childNode.placeOffset = 0
		return childNodeIndex
	}
}

func (t *Quadtree[T]) refresh() {
	if t.isDirty {
		t.groupItems()
		t.updateIDMappings()
		t.updateAABB(0)
		t.isDirty = false
	}
}

func (t *Quadtree[T]) groupItems() {
	offset := uint32(0)
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemOffset = offset
		node.placeOffset = offset
		offset += node.itemCount
	}
	countActiveItems := uint32(offset)

	nullOffset := countActiveItems
	for i := uint32(0); i < countActiveItems; {
		item := &t.items[i]
		if item.node == nullQuadtreeIndex {
			t.swapItems(i, nullOffset)
			nullOffset++
			continue
		}
		node := &t.nodes[item.node]
		if i >= node.itemOffset && i < node.placeOffset {
			i++ // item is in the right place
			continue
		}
		t.swapItems(i, node.placeOffset)
		node.placeOffset++
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
	parentNode.isDirty = true // ensure it updates its AABB
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
			result = mergeQuadtreeAABBs(result, child.tightArea)
		}
	}
	itemIndex := node.itemOffset
	for range node.itemCount {
		item := &t.items[itemIndex]
		result = mergeQuadtreeAABBs(result, item.tightArea)
		itemIndex++
	}
	node.tightArea = result
	node.isDirty = false

	return true
}

func (t *Quadtree[T]) visitNodeInSegment(nodeIndex int32, querySegment *shape2d.Segment, yield VisitorFunc[T]) bool {
	node := &t.nodes[nodeIndex]
	if node.tightArea.intersectsSegment(querySegment) {
		t.nodeCountAccepted++
		itemIndex := node.itemOffset
		for range node.itemCount {
			item := &t.items[itemIndex]
			if item.tightArea.intersectsSegment(querySegment) {
				t.itemCountAccepted++
				if !yield(item.value) {
					return false
				}
			} else {
				t.itemCountRejected++
			}
			itemIndex++
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

func (t *Quadtree[T]) visitNodeInAABB(nodeIndex int32, queryAABB *shape2d.AABB, yield VisitorFunc[T]) bool {
	node := &t.nodes[nodeIndex]
	if node.tightArea.intersectsAABB(queryAABB) {
		t.nodeCountAccepted++
		itemIndex := node.itemOffset
		for range node.itemCount {
			item := &t.items[itemIndex]
			if item.tightArea.intersectsAABB(queryAABB) {
				t.itemCountAccepted++
				if !yield(item.value) {
					return false
				}
			} else {
				t.itemCountRejected++
			}
			itemIndex++
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
	parent   int32
	children [4]int32

	// looseArea is the fixed square that determines which items can be placed
	// in this node. It is twice the size of the node's share of the tree.
	looseArea quadtreeQuad

	// tightArea is the cached bounding box of everything actually stored in
	// this node and its descendants. It is what queries are tested against.
	tightArea quadtreeAABB

	itemCount   uint32
	itemOffset  uint32
	placeOffset uint32
	isDirty     bool
}

func (n *quadtreeNode) isEmpty() bool {
	return (n.children == emptyQuadtreeNodeChildren) && (n.itemCount == 0)
}

type quadtreeItem[T any] struct {
	id        TreeItemID
	node      int32
	tightArea quadtreeAABB
	value     T
}

// quadtreeQuad is a square, described through its center and half-size. It
// describes the loose area of a node, which is what an item has to fit into in
// order to be placed there.
type quadtreeQuad struct {
	x        float64
	y        float64
	halfSize float64
}

type quadtreeAABB struct {
	minX float64
	minY float64
	maxX float64
	maxY float64
}

func emptyQuadtreeAABB() quadtreeAABB {
	return quadtreeAABB{
		minX: math.MaxFloat64,
		minY: math.MaxFloat64,
		maxX: -math.MaxFloat64,
		maxY: -math.MaxFloat64,
	}
}

func newQuadtreeAABBFromAABB(aabb shape2d.AABB) quadtreeAABB {
	return quadtreeAABB{
		minX: aabb.MinX,
		minY: aabb.MinY,
		maxX: aabb.MaxX,
		maxY: aabb.MaxY,
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

func (aabb *quadtreeAABB) intersectsSegment(segment *shape2d.Segment) bool {
	if aabb.isEmpty() {
		return false
	}

	delta := dprec.Vec2Diff(segment.B, segment.A)

	var tCloseX, tFarX float64
	if delta.X == 0.0 {
		if (segment.A.X < aabb.minX) || (segment.A.X > aabb.maxX) {
			return false // both points are outside the box on the left or right
		}
		tCloseX = -math.MaxFloat64
		tFarX = math.MaxFloat64
	} else {
		tLowX := (aabb.minX - segment.A.X) / delta.X
		tHighX := (aabb.maxX - segment.A.X) / delta.X
		tCloseX = min(tLowX, tHighX)
		tFarX = max(tLowX, tHighX)
	}

	var tCloseY, tFarY float64
	if delta.Y == 0.0 {
		if (segment.A.Y < aabb.minY) || (segment.A.Y > aabb.maxY) {
			return false // both points are outside the box on the top or bottom
		}
		tCloseY = -math.MaxFloat64
		tFarY = math.MaxFloat64
	} else {
		tLowY := (aabb.minY - segment.A.Y) / delta.Y
		tHighY := (aabb.maxY - segment.A.Y) / delta.Y
		tCloseY = min(tLowY, tHighY)
		tFarY = max(tLowY, tHighY)
	}

	tClose := max(tCloseX, tCloseY)
	tFar := min(tFarX, tFarY)

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}

func (aabb *quadtreeAABB) intersectsAABB(other *shape2d.AABB) bool {
	if aabb.isEmpty() {
		return false
	}
	return (aabb.minX <= other.MaxX) &&
		(aabb.minY <= other.MaxY) &&
		(aabb.maxX >= other.MinX) &&
		(aabb.maxY >= other.MinY)
}
