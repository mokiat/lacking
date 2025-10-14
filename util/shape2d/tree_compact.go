package shape2d

import (
	"log/slog"
	"slices"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
)

// InvalidCompactTreeItemID is an identifier that can be used by user
// code to mark an identifier as invalid. However, such an identifier will
// never be returned by the library and must also never be passed to the
// library.
const InvalidCompactTreeItemID = CompactTreeItemID(0xFFFFFFFF)

// CompactTreeItemID is an identifier used to control the placement of an item
// into a compact tree.
type CompactTreeItemID uint32

// CompactTreeStats represents the current state of a CompactTree.
type CompactTreeStats struct {

	// NodeCount is the total number of nodes in the tree.
	NodeCount uint32

	// ItemCount is the total number of items in the tree.
	ItemCount uint32

	// ItemCountPerDepth contains the number of items at each depth level.
	ItemCountPerDepth []uint32
}

// CompactTreeVisitStats represents statistics on the last visit operation
// performed on a CompactTree.
type CompactTreeVisitStats struct {

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

// CompactTreeSettings contains the settings for a CompactTree.
type CompactTreeSettings struct {

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

// NewCompactTree creates a new CompactTree using the provided settings.
func NewCompactTree[T any](settings CompactTreeSettings) *CompactTree[T] {
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

	nodes := make([]compactTreeNode, 0, initialNodeCapacity)
	nodes = append(nodes, compactTreeNode{
		parent:    unspecifiedIndex,
		children:  emptyCompactTreeNodeChildren,
		itemStart: 0,
		itemEnd:   0,
		looseArea: CompactQuad{
			x: 0.0,
			y: 0.0,
			r: float32(size), // using size here since a loose cube has twice the radius
		},
		box: emptyCompactAABB(),
	})

	return &CompactTree[T]{
		nodes:           nodes,
		items:           make([]compactTreeItem[T], 0, initialItemCapacity),
		freeNodeIndices: ds.NewStack[int32](32),
		freeItemIDs:     ds.NewStack[CompactTreeItemID](32),
		idMappings:      make([]int32, 0, initialItemCapacity),
		maxDepth:        maxDepth,

		nodeCountAccepted: 0,
		nodeCountRejected: 0,
		itemCountAccepted: 0,
		itemCountRejected: 0,

		isDirty: false,
	}
}

// CompactTree is a spatial structure that uses a loose quadtree implementation
// with shrinking bounding box to enable the fast searching of items.
type CompactTree[T any] struct {
	nodes           []compactTreeNode
	items           []compactTreeItem[T]
	freeNodeIndices *ds.Stack[int32]
	freeItemIDs     *ds.Stack[CompactTreeItemID]
	idMappings      []int32
	maxDepth        uint32

	nodeCountAccepted uint32
	nodeCountRejected uint32
	itemCountAccepted uint32
	itemCountRejected uint32

	isDirty bool
}

// Stats returns statistics on the current state of this tree.
func (t *CompactTree[T]) Stats() CompactTreeStats {
	t.refresh() // this is necessary
	itemCountPerDepth := make([]uint32, t.maxDepth)
	for i := range uint32(t.maxDepth) {
		itemCountPerDepth[i] = t.itemsAtDepth(0, 1, i+1)
	}
	return CompactTreeStats{
		NodeCount:         t.activeNodeCount(),
		ItemCount:         t.activeItemCount(),
		ItemCountPerDepth: itemCountPerDepth,
	}
}

// VisitStats returns statistics information on the last executed search in
// this tree.
func (t *CompactTree[T]) VisitStats() CompactTreeVisitStats {
	return CompactTreeVisitStats{
		NodeCountVisited:  t.nodeCountAccepted + t.nodeCountRejected,
		NodeCountAccepted: t.nodeCountAccepted,
		NodeCountRejected: t.nodeCountRejected,
		ItemCountVisited:  t.itemCountAccepted + t.itemCountRejected,
		ItemCountAccepted: t.itemCountAccepted,
		ItemCountRejected: t.itemCountRejected,
	}
}

// Insert adds an item, which occupies the specified quad area, to this
// tree.
func (t *CompactTree[T]) Insert(quad CompactQuad, value T) CompactTreeItemID {
	t.isDirty = true
	if !t.freeItemIDs.IsEmpty() {
		id := t.freeItemIDs.Pop()
		itemIndex := t.idMappings[id]
		item := &t.items[itemIndex]
		item.box = compactAABBFromQuad(quad)
		item.value = value
		item.node = t.pickNodeForItem(quad)
		return item.id
	} else {
		if len(t.items) == cap(t.items) {
			logger.Warn("Will grow item capacity for compact tree.",
				slog.Int("current", len(t.items)),
			)
		}
		id := CompactTreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		t.items = append(t.items, compactTreeItem[T]{
			id:    id,
			node:  t.pickNodeForItem(quad),
			box:   compactAABBFromQuad(quad),
			value: value,
		})
		return id
	}
}

// Update repositions the item with the specified id to the new area.
func (t *CompactTree[T]) Update(id CompactTreeItemID, quad CompactQuad) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.box = compactAABBFromQuad(quad)
	item.node = t.pickNodeForItem(quad)
}

// Remove removes the item with the specified id from this tree.
func (t *CompactTree[T]) Remove(id CompactTreeItemID) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == unspecifiedIndex {
		panic("cannot remove item twice")
	}
	item.node = unspecifiedIndex
	t.freeItemIDs.Push(id)
}

// QuerySegment finds all items that intersect the specified segment. Each
// found item is passed to the specified yield function. The order in which
// items are passed is undefined and might change between invocations.
func (t *CompactTree[T]) QuerySegment(querySegment CompactQuerySegment, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInSegment(0, querySegment, yield)
}

// QueryAABB finds all items that are inside or intersect the specified
// axis-aligned bounding box. Each found item is passed to the specified yield
// function. The order in which items are passed is undefined and might change
// between invocations.
func (t *CompactTree[T]) QueryAABB(queryBox CompactQueryAABB, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	box := compactAABB(queryBox)
	t.visitNodeInAABB(0, &box, yield)
}

// GC runs cleanup and optimization logic. You should call this at least once
// per frame.
func (t *CompactTree[T]) GC() {
	t.refresh()
}

func (t *CompactTree[T]) resetVisitStats() {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
}

func (t *CompactTree[T]) activeNodeCount() uint32 {
	return uint32(len(t.nodes) - t.freeNodeIndices.Size())
}

func (t *CompactTree[T]) activeItemCount() uint32 {
	return uint32(len(t.items) - t.freeItemIDs.Size())
}

func (t *CompactTree[T]) itemsAtDepth(nodeIndex int32, currentDepth, depth uint32) uint32 {
	if nodeIndex == unspecifiedIndex {
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

func (t *CompactTree[T]) pickNodeForItem(quad CompactQuad) int32 {
	bestNodeIndex := unspecifiedIndex
	currentNodeIndex := int32(0)
	var depth uint32
	for currentNodeIndex != unspecifiedIndex {
		bestNodeIndex = currentNodeIndex
		depth++
		if depth >= t.maxDepth {
			break
		}
		currentNodeIndex = t.pickChildNode(currentNodeIndex, quad)
	}
	return bestNodeIndex
}

func (t *CompactTree[T]) pickChildNode(parentNodeIndex int32, quad CompactQuad) int32 {
	parentNode := &t.nodes[parentNodeIndex]
	parentLooseArea := parentNode.looseArea

	// Make sure that it can fit inside a child. The requirement is that
	// the radius must be smaller than the loose margin of the child.
	childLooseRadius := parentLooseArea.r / 2.0
	if quad.r > (childLooseRadius / 2.0) { // div by 2 to convert to margin
		return unspecifiedIndex
	}

	// It has to be inside one of the four children.
	var (
		childIndex = 0
		childX     = parentLooseArea.x
		childY     = parentLooseArea.y
	)
	childOffset := parentLooseArea.r / 4.0
	if quad.x < parentLooseArea.x {
		childX -= childOffset
	} else {
		childIndex += 1
		childX += childOffset
	}
	if quad.y < parentLooseArea.y {
		childY -= childOffset
	} else {
		childIndex += 2
		childY += childOffset
	}

	if parentNode.children[childIndex] != unspecifiedIndex {
		return parentNode.children[childIndex]
	}

	childLooseArea := CompactQuad{
		x: childX,
		y: childY,
		r: childLooseRadius,
	}
	if !t.freeNodeIndices.IsEmpty() {
		childNodeIndex := t.freeNodeIndices.Pop()
		parentNode.children[childIndex] = childNodeIndex
		childNode := &t.nodes[childNodeIndex]
		childNode.parent = parentNodeIndex
		childNode.children = emptyCompactTreeNodeChildren
		childNode.looseArea = childLooseArea
		childNode.itemStart = 0
		childNode.itemEnd = 0
		return childNodeIndex
	} else {
		if len(t.nodes) == cap(t.nodes) {
			logger.Warn("Will grow node capacity for compact tree.",
				slog.Int("current", len(t.nodes)),
			)
		}
		childNodeIndex := int32(len(t.nodes)) // predict next node index
		parentNode.children[childIndex] = childNodeIndex
		// Do NOT use "parentNode" after this append as the ref might be towards
		// an old slice!
		t.nodes = append(t.nodes, compactTreeNode{
			parent:    parentNodeIndex,
			children:  emptyCompactTreeNodeChildren,
			looseArea: childLooseArea,
			itemStart: 0,
			itemEnd:   0,
		})
		return childNodeIndex
	}
}

func (t *CompactTree[T]) refresh() {
	if t.isDirty {
		t.sortItems()
		t.eraseItemOffsets()
		t.evaluateItemOffsets()
		t.updateIDMappings()
		t.gcNodes()
		t.updateAABB(0)
		t.isDirty = false
	}
}

func (t *CompactTree[T]) sortItems() {
	// TODO: Test if this can be made faster with ref sorting.
	slices.SortFunc(t.items, compareCompactTreeItems[T])
}

func (t *CompactTree[T]) eraseItemOffsets() {
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart = 0
		node.itemEnd = 0
	}
}

func (t *CompactTree[T]) evaluateItemOffsets() {
	lastNode := unspecifiedIndex
	itemIndex := uint32(0)
	itemCount := uint32(len(t.items))
	for itemIndex < itemCount {
		item := &t.items[itemIndex]
		if item.node != lastNode {
			if lastNode != unspecifiedIndex {
				t.nodes[lastNode].itemEnd = itemIndex
			}
			t.nodes[item.node].itemStart = itemIndex
		}
		lastNode = item.node
		itemIndex++
	}
	if lastNode != unspecifiedIndex {
		t.nodes[lastNode].itemEnd = itemIndex
	}
}

func (t *CompactTree[T]) updateIDMappings() {
	for i, item := range t.items {
		t.idMappings[item.id] = int32(i)
	}
}

func (t *CompactTree[T]) gcNodes() {
	for i := range t.nodes {
		t.gcNode(int32(i))
	}
}

func (t *CompactTree[T]) gcNode(nodeIndex int32) {
	node := &t.nodes[nodeIndex]
	if node.parent == unspecifiedIndex {
		return // already deleted or root
	}
	if !node.isEmpty() {
		return // can't gc node
	}
	parentNodeIndex := node.parent
	parentNode := &t.nodes[parentNodeIndex]
	for i, childNodeIndex := range parentNode.children {
		if childNodeIndex == nodeIndex {
			parentNode.children[i] = unspecifiedIndex
			break
		}
	}
	node.parent = unspecifiedIndex
	t.freeNodeIndices.Push(nodeIndex)
	t.gcNode(parentNodeIndex)
}

func (t *CompactTree[T]) updateAABB(nodeIndex int32) compactAABB {
	node := &t.nodes[nodeIndex]

	// The AABB is created flipped so that the first box to be merged will
	// override this completely. Also, even if it is not overridden, it will
	// not match anything in this initial form.
	result := emptyCompactAABB()

	for _, childIndex := range node.children {
		if childIndex != unspecifiedIndex {
			childBox := t.updateAABB(childIndex)
			result = mergeCompactAABBs(result, childBox)
		}
	}

	for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
		item := &t.items[itemIndex]
		result = mergeCompactAABBs(result, item.box)
	}

	node.box = result
	return result
}

func (t *CompactTree[T]) visitNodeInAABB(nodeIndex int32, box *compactAABB, yield VisitorFunc[T]) {
	node := &t.nodes[nodeIndex]
	if box.intersects(node.box) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if box.intersects(item.box) {
				if !yield(item.value) {
					return
				}
				t.itemCountAccepted++
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			if childNodeIndex != unspecifiedIndex {
				t.visitNodeInAABB(childNodeIndex, box, yield)
			}
		}
	} else {
		t.nodeCountRejected++
	}
}

func (t *CompactTree[T]) visitNodeInSegment(nodeIndex int32, segment CompactQuerySegment, yield VisitorFunc[T]) {
	node := &t.nodes[nodeIndex]
	if isCompactSegmentAABBIntersection(segment, node.box) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if isCompactSegmentAABBIntersection(segment, item.box) {
				if !yield(item.value) {
					return
				}
				t.itemCountAccepted++
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			if childNodeIndex != unspecifiedIndex {
				t.visitNodeInSegment(childNodeIndex, segment, yield)
			}
		}
	} else {
		t.nodeCountRejected++
	}
}

// NewCompactQuerySegment creates a new CompactQuerySegment instance from the
// specified endpoints.
func NewCompactQuerySegment(a, b dprec.Vec2) CompactQuerySegment {
	return CompactQuerySegment{
		a: dtos.Vec2(a),
		b: dtos.Vec2(b),
	}
}

// CompactQuerySegment represents a line segment used for querying the tree.
type CompactQuerySegment struct {
	a sprec.Vec2
	b sprec.Vec2
}

// NewCompactQueryAABB creates a new CompactQueryAABB instance from the
// specified bounds.
func NewCompactQueryAABB(minX, maxX, minY, maxY float64) CompactQueryAABB {
	return CompactQueryAABB{
		minX: float32(minX),
		maxX: float32(maxX),
		minY: float32(minY),
		maxY: float32(maxY),
	}
}

// NewCompactQueryAABBFromCircle creates a new CompactQueryAABB that wraps a
// circle.
func NewCompactQueryAABBFromCircle(position dprec.Vec2, radius float64) CompactQueryAABB {
	return CompactQueryAABB{
		minX: float32(position.X - radius),
		maxX: float32(position.X + radius),
		minY: float32(position.Y - radius),
		maxY: float32(position.Y + radius),
	}
}

// CompactQueryAABB represents an axis-aligned bounding box used for querying
// the tree.
type CompactQueryAABB struct {
	minX float32
	maxX float32
	minY float32
	maxY float32
}

// NewCompactQuad creates a new CompactQuad instance from the specified
// position and size.
func NewCompactQuad(x, y, size float64) CompactQuad {
	return CompactQuad{
		x: float32(x),
		y: float32(y),
		r: float32(size / 2.0),
	}
}

// NewCompactQuadFromCircle creates a new CompactQuad that wraps a circle.
func NewCompactQuadFromCircle(position dprec.Vec2, radius float64) CompactQuad {
	return CompactQuad{
		x: float32(position.X),
		y: float32(position.Y),
		r: float32(radius),
	}
}

// CompactQuad represents a square area used for inserting items into the tree.
type CompactQuad struct {
	x float32
	y float32
	r float32
}

const unspecifiedIndex = int32(-1)

var emptyCompactTreeNodeChildren = [4]int32{
	unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
}

type compactTreeNode struct {
	parent    int32
	children  [4]int32
	looseArea CompactQuad
	box       compactAABB
	itemStart uint32
	itemEnd   uint32
}

func (n *compactTreeNode) isEmpty() bool {
	// TODO: Check if emptyCompactTreeNodeChildren compare is faster.
	// if n.children != emptyCompactTreeNodeChildren {
	// 	return false
	// }
	for _, childIndex := range n.children {
		if childIndex != unspecifiedIndex {
			return false
		}
	}
	return n.itemStart >= n.itemEnd
}

type compactTreeItem[T any] struct {
	id    CompactTreeItemID
	node  int32
	box   compactAABB
	value T
}

func compareCompactTreeItems[T any](a, b compactTreeItem[T]) int {
	return int(a.node - b.node)
}

func emptyCompactAABB() compactAABB {
	const large = 128000.0
	return compactAABB{
		minX: large,
		maxX: -large,
		minY: large,
		maxY: -large,
	}
}

func mergeCompactAABBs(first compactAABB, second compactAABB) compactAABB {
	return compactAABB{
		minX: min(first.minX, second.minX),
		maxX: max(first.maxX, second.maxX),
		minY: min(first.minY, second.minY),
		maxY: max(first.maxY, second.maxY),
	}
}

func compactAABBFromQuad(area CompactQuad) compactAABB {
	return compactAABB{
		minX: area.x - area.r,
		maxX: area.x + area.r,
		minY: area.y - area.r,
		maxY: area.y + area.r,
	}
}

type compactAABB struct {
	minX float32
	maxX float32
	minY float32
	maxY float32
}

func (box compactAABB) intersects(other compactAABB) bool {
	return (box.minX <= other.maxX) &&
		(box.maxX >= other.minX) &&
		(box.minY <= other.maxY) &&
		(box.maxY >= other.minY)
}

// TODO: Test if passing references is faster.
func isCompactSegmentAABBIntersection(segment CompactQuerySegment, aabb compactAABB) bool {
	delta := sprec.Vec2Diff(segment.b, segment.a)

	tLowX := (aabb.minX - segment.a.X) / delta.X
	tLowY := (aabb.minY - segment.a.Y) / delta.Y

	tHighX := (aabb.maxX - segment.a.X) / delta.X
	tHighY := (aabb.maxY - segment.a.Y) / delta.Y

	tCloseX := min(tLowX, tHighX)
	tCloseY := min(tLowY, tHighY)
	tClose := max(tCloseX, tCloseY)

	tFarX := max(tLowX, tHighX)
	tFarY := max(tLowY, tHighY)
	tFar := min(tFarX, tFarY)

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}
