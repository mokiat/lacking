package spatial

import (
	"log/slog"
	"slices"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
)

// InvalidCompactOctreeItemID can be used to mark a reference as invalid.
const InvalidCompactOctreeItemID = CompactOctreeItemID(0xFFFFFFFF)

// CompactOctreeItemID is an identifier used to control the placement of an item
// into a compact octree.
type CompactOctreeItemID uint32

// CompactOctreeStats represents the current state of a CompactOctree.
type CompactOctreeStats struct {
	NodeCount          int32
	ItemCount          int32
	ItemsCountPerDepth []int32
}

// CompactOctreeVisitStats represents statistics on the last visit operation
// performed on a CompactOctree.
type CompactOctreeVisitStats struct {
	NodeCount         int32
	NodeCountVisited  int32
	NodeCountAccepted int32
	NodeCountRejected int32
	ItemCount         int32
	ItemCountVisited  int32
	ItemCountAccepted int32
	ItemCountRejected int32
}

// CompactOctreeSettings contains the settings for a CompactOctree.
type CompactOctreeSettings struct {

	// Size specifies the dimension (side to side) of the octree cube.
	// If not specified, a default size of 4096 is used.
	// Inserting an item outside these bounds has undefined behavior.
	Size opt.T[float64]

	// MaxDepth controls the maximum depth that the octree can reach.
	// If not specified, a default max depth of 8 is used.
	MaxDepth opt.T[int32]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. Usually one would find that number empirically.
	// This allows the octree to preallocate memory and avoid dynamic allocations.
	//
	// By default the initial capacity is 4096.
	InitialNodeCapacity opt.T[int32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the octree. This allows the octree to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[int32]
}

// NewCompactOctree creates a new CompactOctree using the provided settings.
func NewCompactOctree[T any](settings CompactOctreeSettings) *CompactOctree[T] {
	size := settings.Size.ValueOrDefault(4096.0)
	if size < 1.0 {
		panic("size cannot be smaller than 1")
	}
	maxDepth := settings.MaxDepth.ValueOrDefault(8)
	if maxDepth < 1 {
		panic("max depth cannot be smaller than 1")
	}
	initialNodeCapacity := settings.InitialNodeCapacity.ValueOrDefault(4096)
	if initialNodeCapacity < 0 {
		panic("initial node capacity must not be negative")
	}
	initialItemCapacity := settings.InitialItemCapacity.ValueOrDefault(1024)
	if initialItemCapacity < 0 {
		panic("initial item capacity must not be negative")
	}

	nodes := make([]compactOctreeNode, 0, initialNodeCapacity)
	nodes = append(nodes, compactOctreeNode{
		parent: unspecifiedIndex,
		children: [8]int32{
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
		},
		itemStart: unspecifiedIndex,
		itemEnd:   unspecifiedIndex,
		looseArea: CubeArea{
			x: 0.0,
			y: 0.0,
			z: 0.0,
			r: size, // using size here since a loose cube has twice the radius
		},
		box: compactAABB{},
	})
	items := make([]compactOctreeItem[T], 0, initialItemCapacity)
	idMappings := make([]int32, 0, initialItemCapacity)

	return &CompactOctree[T]{
		nodes:           nodes,
		items:           items,
		freeNodeIndices: ds.NewStack[int32](32),
		freeItemIndices: ds.NewStack[int32](32),
		idMappings:      idMappings,
		maxDepth:        maxDepth,
	}
}

// CompactOctree is a spatial structure that uses a loose octree implementation
// with biased placement to enable the fast search of items within a region.
type CompactOctree[T any] struct {
	nodes           []compactOctreeNode
	items           []compactOctreeItem[T]
	freeNodeIndices *ds.Stack[int32]
	freeItemIndices *ds.Stack[int32]
	idMappings      []int32
	maxDepth        int32

	nodeCountAccepted int32
	nodeCountRejected int32
	itemCountAccepted int32
	itemCountRejected int32

	isDirty bool
}

// Stats returns statistics on the current state of this octree.
func (t *CompactOctree[T]) Stats() CompactOctreeStats {
	itemCountPerDepth := make([]int32, t.maxDepth+1)
	for i := int32(1); i <= t.maxDepth; i++ {
		itemCountPerDepth[i] = t.itemsAtDepth(0, 1, i)
	}
	return CompactOctreeStats{
		NodeCount:          int32(len(t.nodes)),
		ItemCount:          int32(len(t.items)),
		ItemsCountPerDepth: itemCountPerDepth,
	}
}

// VisitStats returns statistics information on the last executed search in
// this octree.
func (t *CompactOctree[T]) VisitStats() CompactOctreeVisitStats {
	return CompactOctreeVisitStats{
		NodeCount:         int32(len(t.nodes) - t.freeNodeIndices.Size()),
		NodeCountVisited:  t.nodeCountAccepted + t.nodeCountRejected,
		NodeCountAccepted: t.nodeCountAccepted,
		NodeCountRejected: t.nodeCountRejected,
		ItemCount:         int32(len(t.items) - t.freeItemIndices.Size()),
		ItemCountVisited:  t.itemCountAccepted + t.itemCountRejected,
		ItemCountAccepted: t.itemCountAccepted,
		ItemCountRejected: t.itemCountRejected,
	}
}

// Insert adds an item to this octree at the specified position and taking the
// specified radius into account.
func (t *CompactOctree[T]) Insert(area CubeArea, value T) CompactOctreeItemID {
	t.isDirty = true
	if !t.freeItemIndices.IsEmpty() {
		itemIndex := t.freeItemIndices.Pop()
		item := &t.items[itemIndex]
		item.box = compactAABBFromCube(area)
		item.value = value
		item.node = t.pickNodeForItem(area)
		return item.id
	} else {
		if len(t.items) == cap(t.items) {
			logger.Warn("Will grow item capacity for compact octree.",
				slog.Int("current", len(t.items)),
			)
		}
		id := CompactOctreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		t.items = append(t.items, compactOctreeItem[T]{
			id:    id,
			node:  t.pickNodeForItem(area),
			box:   compactAABBFromCube(area),
			value: value,
		})
		return id
	}
}

// Update repositions the item with the specified id to the new area.
func (t *CompactOctree[T]) Update(id CompactOctreeItemID, area CubeArea) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.box = compactAABBFromCube(area)
	item.node = t.pickNodeForItem(area)
}

// Remove removes the item with the specified id from this data structure.
func (t *CompactOctree[T]) Remove(id CompactOctreeItemID) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.node = unspecifiedIndex
	t.freeItemIndices.Push(itemIndex)
}

// VisitArea finds all items that are inside or intersect the specified area.
// It calls the specified visitor for each item found.
func (t *CompactOctree[T]) VisitArea(area CubeArea, visitor Visitor[T]) {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
	t.refresh()
	box := compactAABBFromCube(area)
	t.visitNodeInAABB(0, &box, visitor)
}

// VisitSegment finds all items that are inside or intersect the specified
// segment.
// It calls the specified visitor for each item found.
func (t *CompactOctree[T]) VisitSegment(a, b dprec.Vec3, visitor Visitor[T]) {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
	t.refresh()
	segment := compactSegment{
		a: dtos.Vec3(a),
		b: dtos.Vec3(b),
	}
	t.visitNodeInSegment(0, segment, visitor)
}

// GC runs cleanup and optimization logic. You should call this at least once
// per frame.
func (t *CompactOctree[T]) GC() {
	t.refresh()
}

func (t *CompactOctree[T]) pickNodeForItem(area CubeArea) int32 {
	bestNodeIndex := unspecifiedIndex
	currentNodeIndex := int32(0)
	depth := int32(0)
	for currentNodeIndex != unspecifiedIndex {
		bestNodeIndex = currentNodeIndex
		depth++
		if depth >= t.maxDepth {
			break
		}
		currentNodeIndex = t.pickChildNode(currentNodeIndex, area)
	}
	return bestNodeIndex
}

func (t *CompactOctree[T]) pickChildNode(parentNodeIndex int32, area CubeArea) int32 {
	parentNode := &t.nodes[parentNodeIndex]
	parentLooseArea := parentNode.looseArea

	// Make sure that it can fit inside a child. The requirement is that
	// the radius must be smaller than the loose margin of the child.
	childLooseRadius := parentLooseArea.r / 2.0
	if area.r > (childLooseRadius / 2.0) { // div by 2 to convert to margin
		return unspecifiedIndex
	}

	// It has to be inside one of the eight children.
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

	if parentNode.children[childIndex] != unspecifiedIndex {
		return parentNode.children[childIndex]
	}

	childLooseArea := CubeArea{
		x: childX,
		y: childY,
		z: childZ,
		r: childLooseRadius,
	}
	if !t.freeNodeIndices.IsEmpty() {
		childNodeIndex := t.freeNodeIndices.Pop()
		parentNode.children[childIndex] = childNodeIndex
		childNode := &t.nodes[childNodeIndex]
		childNode.parent = parentNodeIndex
		childNode.children = [8]int32{
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
		}
		childNode.looseArea = childLooseArea
		childNode.itemStart = 0
		childNode.itemEnd = 0
		return childNodeIndex
	} else {
		if len(t.nodes) == cap(t.nodes) {
			logger.Warn("Will grow node capacity for compact octree.",
				slog.Int("current", len(t.nodes)),
			)
		}
		childNodeIndex := int32(len(t.nodes)) // predict next node index
		parentNode.children[childIndex] = childNodeIndex
		// NOTE: Do NOT use parentNode after this append as the ref might be towards
		// an old slice.
		t.nodes = append(t.nodes, compactOctreeNode{
			parent: parentNodeIndex,
			children: [8]int32{
				unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
				unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
			},
			looseArea: childLooseArea,
			itemStart: unspecifiedIndex,
			itemEnd:   unspecifiedIndex,
		})
		return childNodeIndex
	}
}

func (t *CompactOctree[T]) visitNodeInAABB(nodeIndex int32, box *compactAABB, visitor Visitor[T]) {
	node := &t.nodes[nodeIndex]
	if box.intersects(node.box) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if box.intersects(item.box) {
				visitor.Visit(item.value)
				t.itemCountAccepted++
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			if childNodeIndex != unspecifiedIndex {
				t.visitNodeInAABB(childNodeIndex, box, visitor)
			}
		}
	} else {
		t.nodeCountRejected++
	}
}

func (t *CompactOctree[T]) visitNodeInSegment(nodeIndex int32, segment compactSegment, visitor Visitor[T]) {
	node := &t.nodes[nodeIndex]
	if isCompactSegmentAABBIntersection(segment, node.box) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if isCompactSegmentAABBIntersection(segment, item.box) {
				visitor.Visit(item.value)
				t.itemCountAccepted++
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			if childNodeIndex != unspecifiedIndex {
				t.visitNodeInSegment(childNodeIndex, segment, visitor)
			}
		}
	} else {
		t.nodeCountRejected++
	}
}

func (t *CompactOctree[T]) refresh() {
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

func (t *CompactOctree[T]) sortItems() {
	// TODO: Test if this can be made faster with ref sorting.
	slices.SortFunc(t.items, compareCompactOctreeItems[T])
}

func (t *CompactOctree[T]) eraseItemOffsets() {
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart = unspecifiedIndex
		node.itemEnd = unspecifiedIndex
	}
}

func (t *CompactOctree[T]) evaluateItemOffsets() {
	lastNode := unspecifiedIndex
	itemIndex := int32(0)
	itemCount := int32(len(t.items))
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

func (t *CompactOctree[T]) updateIDMappings() {
	for i, item := range t.items {
		t.idMappings[item.id] = int32(i)
	}
}

func (t *CompactOctree[T]) gcNodes() {
	for i := range t.nodes {
		t.gcNode(int32(i))
	}
}

func (t *CompactOctree[T]) gcNode(nodeIndex int32) {
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

func (t *CompactOctree[T]) updateAABB(nodeIndex int32) compactAABB {
	node := &t.nodes[nodeIndex]

	// The AABB is created flipped so that the first box to be merged will
	// override this completely. Also, even if it is not overridden, it will
	// not match anything in this initial form.
	const large = 128000.0
	result := compactAABB{
		minX: large,
		maxX: -large,
		minY: large,
		maxY: -large,
		minZ: large,
		maxZ: -large,
	}

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

func (t *CompactOctree[T]) itemsAtDepth(nodeIndex, currentDepth, depth int32) int32 {
	if nodeIndex == unspecifiedIndex {
		return 0
	}
	node := &t.nodes[nodeIndex]
	if currentDepth == depth {
		return node.itemEnd - node.itemStart
	}
	result := int32(0)
	for _, childNodeIndex := range node.children {
		result += t.itemsAtDepth(childNodeIndex, currentDepth+1, depth)
	}
	return result
}

func compareCompactOctreeItems[T any](a, b compactOctreeItem[T]) int {
	return int(a.node - b.node)
}

type compactOctreeNode struct {
	parent    int32
	children  [8]int32
	looseArea CubeArea
	box       compactAABB
	itemStart int32
	itemEnd   int32
}

func (n *compactOctreeNode) isEmpty() bool {
	for _, childIndex := range n.children {
		if childIndex != unspecifiedIndex {
			return false
		}
	}
	return n.itemStart >= n.itemEnd
}

type compactOctreeItem[T any] struct {
	id    CompactOctreeItemID
	node  int32
	box   compactAABB
	value T
}

func mergeCompactAABBs(first compactAABB, second compactAABB) compactAABB {
	return compactAABB{
		minX: min(first.minX, second.minX),
		maxX: max(first.maxX, second.maxX),
		minY: min(first.minY, second.minY),
		maxY: max(first.maxY, second.maxY),
		minZ: min(first.minZ, second.minZ),
		maxZ: max(first.maxZ, second.maxZ),
	}
}

func compactAABBFromCube(area CubeArea) compactAABB {
	return compactAABB{
		minX: float32(area.x - area.r),
		maxX: float32(area.x + area.r),
		minY: float32(area.y - area.r),
		maxY: float32(area.y + area.r),
		minZ: float32(area.z - area.r),
		maxZ: float32(area.z + area.r),
	}
}

type compactAABB struct {
	minX float32
	maxX float32
	minY float32
	maxY float32
	minZ float32
	maxZ float32
}

func (box compactAABB) intersects(other compactAABB) bool {
	return (box.minX <= other.maxX) && (box.maxX >= other.minX) &&
		(box.minY <= other.maxY) && (box.maxY >= other.minY) &&
		(box.minZ <= other.maxZ) && (box.maxZ >= other.minZ)
}

type compactSegment struct {
	a sprec.Vec3
	b sprec.Vec3
}

func isCompactSegmentAABBIntersection(segment compactSegment, aabb compactAABB) bool {
	delta := sprec.Vec3Diff(segment.b, segment.a)

	tLowX := (aabb.minX - segment.a.X) / delta.X
	tLowY := (aabb.minY - segment.a.Y) / delta.Y
	tLowZ := (aabb.minZ - segment.a.Z) / delta.Z

	tHighX := (aabb.maxX - segment.a.X) / delta.X
	tHighY := (aabb.maxY - segment.a.Y) / delta.Y
	tHighZ := (aabb.maxZ - segment.a.Z) / delta.Z

	tCloseX := min(tLowX, tHighX)
	tCloseY := min(tLowY, tHighY)
	tCloseZ := min(tLowZ, tHighZ)
	tClose := max(tCloseX, tCloseY, tCloseZ)

	tFarX := max(tLowX, tHighX)
	tFarY := max(tLowY, tHighY)
	tFarZ := max(tLowZ, tHighZ)
	tFar := min(tFarX, tFarY, tFarZ)

	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}
