package shape3d

import (
	"log/slog"

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
		looseArea: compactCube{
			x: 0.0,
			y: 0.0,
			z: 0.0,
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

// CompactTree is a spatial structure that uses a loose octree implementation
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

// Insert adds an item, which occupies the specified cube area, to this
// tree.
func (t *CompactTree[T]) Insert(cube CompactCube, value T) CompactTreeItemID {
	node := t.pickNodeForItem(cube)
	box := compactAABBFromCube(cube)
	t.markNodeDirty(node)

	if t.freeItemIDs.IsEmpty() {
		if len(t.items) == cap(t.items) {
			logger.Warn("Will grow item capacity for compact tree.",
				slog.Int("current", len(t.items)),
			)
		}
		id := CompactTreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		t.items = append(t.items, compactTreeItem[T]{
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
func (t *CompactTree[T]) Update(id CompactTreeItemID, cube CompactCube) {
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.box = compactAABBFromCube(cube)
	t.markNodeDirty(item.node) // previous node
	item.node = t.pickNodeForItem(cube)
	t.markNodeDirty(item.node) // new node
}

// Remove removes the item with the specified id from this tree.
func (t *CompactTree[T]) Remove(id CompactTreeItemID) {
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	if item.node == unspecifiedIndex {
		panic("cannot remove item twice")
	}
	t.markNodeDirty(item.node)
	item.node = unspecifiedIndex
	t.freeItemIDs.Push(id)
}

// QuerySegment finds all items that intersect the specified segment. Each
// found item is passed to the specified yield function. The order in which
// items are passed is undefined and might change between invocations.
func (t *CompactTree[T]) QuerySegment(querySegment CompactQuerySegment, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInSegment(0, &querySegment, yield)
}

// QueryAABB finds all items that are inside or intersect the specified
// axis-aligned bounding box. Each found item is passed to the specified yield
// function. The order in which items are passed is undefined and might change
// between invocations.
func (t *CompactTree[T]) QueryAABB(queryBox CompactQueryAABB, yield VisitorFunc[T]) {
	t.resetVisitStats()
	t.refresh()
	t.visitNodeInAABB(0, &queryBox, yield)
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

func (t *CompactTree[T]) markNodeDirty(nodeIndex int32) {
	t.isDirty = true
	node := &t.nodes[nodeIndex]
	node.isDirty = true
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

func (t *CompactTree[T]) pickNodeForItem(cube CompactCube) int32 {
	bestNodeIndex := unspecifiedIndex
	currentNodeIndex := int32(0)
	var depth uint32
	for currentNodeIndex != unspecifiedIndex {
		bestNodeIndex = currentNodeIndex
		depth++
		if depth >= t.maxDepth {
			break
		}
		currentNodeIndex = t.pickChildNode(currentNodeIndex, cube)
	}
	return bestNodeIndex
}

func (t *CompactTree[T]) pickChildNode(parentNodeIndex int32, cube CompactCube) int32 {
	parentNode := &t.nodes[parentNodeIndex]
	parentLooseArea := parentNode.looseArea

	// Make sure that it can fit inside a child. The requirement is that
	// the radius must be smaller than the loose margin of the child.
	childLooseRadius := parentLooseArea.r / 2.0
	if cube.r > (childLooseRadius / 2.0) { // div by 2 to convert to margin
		return unspecifiedIndex
	}

	// It has to be inside one of the four children.
	var (
		childIndex = 0
		childX     = parentLooseArea.x
		childY     = parentLooseArea.y
		childZ     = parentLooseArea.z
	)
	childOffset := parentLooseArea.r / 4.0
	if cube.x < parentLooseArea.x {
		childX -= childOffset
	} else {
		childIndex += 1
		childX += childOffset
	}
	if cube.z < parentLooseArea.z {
		childZ -= childOffset
	} else {
		childIndex += 2
		childZ += childOffset
	}
	if cube.y < parentLooseArea.y {
		childY -= childOffset
	} else {
		childIndex += 4
		childY += childOffset
	}

	if parentNode.children[childIndex] != unspecifiedIndex {
		return parentNode.children[childIndex]
	}

	childLooseArea := compactCube{
		x: childX,
		y: childY,
		z: childZ,
		r: childLooseRadius,
	}
	if t.freeNodeIndices.IsEmpty() {
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
	} else {
		childNodeIndex := t.freeNodeIndices.Pop()
		parentNode.children[childIndex] = childNodeIndex
		childNode := &t.nodes[childNodeIndex]
		childNode.parent = parentNodeIndex
		childNode.children = emptyCompactTreeNodeChildren
		childNode.looseArea = childLooseArea
		childNode.itemStart = 0
		childNode.itemEnd = 0
		return childNodeIndex
	}
}

func (t *CompactTree[T]) refresh() {
	if t.isDirty {
		t.groupItems()
		t.updateIDMappings()
		t.gcNodes()
		t.updateAABB(0)
		t.isDirty = false
	}
}

func (t *CompactTree[T]) groupItems() {
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart = 0
		node.itemEnd = 0
		node.sortEnd = 0
	}
	for i := range t.items {
		item := &t.items[i]
		if item.node != unspecifiedIndex {
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
		if item.node == unspecifiedIndex {
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

func (t *CompactTree[T]) swapItems(i, j uint32) {
	if i != j {
		t.items[i], t.items[j] = t.items[j], t.items[i]
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

func (t *CompactTree[T]) updateAABB(nodeIndex int32) bool {
	node := &t.nodes[nodeIndex]

	var wereChildrenDirty bool
	for _, childIndex := range node.children {
		if childIndex != unspecifiedIndex {
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

	result := emptyCompactAABB()
	for _, childIndex := range node.children {
		if childIndex != unspecifiedIndex {
			child := &t.nodes[childIndex]
			result = mergeCompactAABBs(result, child.box)
		}
	}
	for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
		item := &t.items[itemIndex]
		result = mergeCompactAABBs(result, item.box)
	}
	node.box = result
	node.isDirty = false

	return true
}

func (t *CompactTree[T]) visitNodeInAABB(nodeIndex int32, box *CompactQueryAABB, yield VisitorFunc[T]) {
	node := &t.nodes[nodeIndex]
	if node.box.intersectsAABB(box) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if item.box.intersectsAABB(box) {
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

func (t *CompactTree[T]) visitNodeInSegment(nodeIndex int32, segment *CompactQuerySegment, yield VisitorFunc[T]) {
	node := &t.nodes[nodeIndex]
	if node.box.intersectsSegment(segment) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if item.box.intersectsSegment(segment) {
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
func NewCompactQuerySegment(a, b dprec.Vec3) CompactQuerySegment {
	return CompactQuerySegment{
		a: dtos.Vec3(a),
		b: dtos.Vec3(b),
	}
}

// CompactQuerySegment represents a line segment used for querying the tree.
type CompactQuerySegment struct {
	a sprec.Vec3
	b sprec.Vec3
}

// NewCompactQueryAABB creates a new CompactQueryAABB instance from the
// specified bounds.
func NewCompactQueryAABB(minX, maxX, minY, maxY, minZ, maxZ float64) CompactQueryAABB {
	return CompactQueryAABB{
		minX: float32(minX),
		maxX: float32(maxX),
		minY: float32(minY),
		maxY: float32(maxY),
		minZ: float32(minZ),
		maxZ: float32(maxZ),
	}
}

// NewCompactQueryAABBFromSphere creates a new CompactQueryAABB that wraps a
// sphere.
func NewCompactQueryAABBFromSphere(position dprec.Vec3, radius float64) CompactQueryAABB {
	return CompactQueryAABB{
		minX: float32(position.X - radius),
		maxX: float32(position.X + radius),
		minY: float32(position.Y - radius),
		maxY: float32(position.Y + radius),
		minZ: float32(position.Z - radius),
		maxZ: float32(position.Z + radius),
	}
}

// CompactQueryAABB represents an axis-aligned bounding box used for querying
// the tree.
type CompactQueryAABB struct {
	minX float32
	maxX float32
	minY float32
	maxY float32
	minZ float32
	maxZ float32
}

// NewCompactCube creates a new CompactCube instance from the specified
// position and size.
func NewCompactCube(x, y, z, size float64) CompactCube {
	return CompactCube{
		x: float32(x),
		y: float32(y),
		z: float32(z),
		r: float32(size / 2.0),
	}
}

// NewCompactCubeFromSphere creates a new CompactCube that wraps a sphere.
func NewCompactCubeFromSphere(position dprec.Vec3, radius float64) CompactCube {
	return CompactCube{
		x: float32(position.X),
		y: float32(position.Y),
		z: float32(position.Z),
		r: float32(radius),
	}
}

// CompactCube represents a cube area used for inserting items into the tree.
type CompactCube struct {
	x float32
	y float32
	z float32
	r float32
}

const unspecifiedIndex = int32(-1)

var emptyCompactTreeNodeChildren = [8]int32{
	unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
	unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
}

type compactTreeNode struct {
	parent    int32
	children  [8]int32
	looseArea compactCube
	box       compactAABB
	itemStart uint32
	itemEnd   uint32
	sortEnd   uint32
	isDirty   bool
}

func (n *compactTreeNode) isEmpty() bool {
	return (n.children == emptyCompactTreeNodeChildren) && (n.itemStart >= n.itemEnd)
}

type compactTreeItem[T any] struct {
	id    CompactTreeItemID
	node  int32
	box   compactAABB
	value T
}

func emptyCompactAABB() compactAABB {
	const large = 128000.0
	return compactAABB{
		minX: large,
		maxX: -large,
		minY: large,
		maxY: -large,
		minZ: large,
		maxZ: -large,
	}
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

func compactAABBFromCube(area CompactCube) compactAABB {
	return compactAABB{
		minX: area.x - area.r,
		maxX: area.x + area.r,
		minY: area.y - area.r,
		maxY: area.y + area.r,
		minZ: area.z - area.r,
		maxZ: area.z + area.r,
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

func (box *compactAABB) intersectsSegment(segment *CompactQuerySegment) bool {
	delta := sprec.Vec3Diff(segment.b, segment.a)

	tLowX := (box.minX - segment.a.X) / delta.X
	tLowY := (box.minY - segment.a.Y) / delta.Y
	tLowZ := (box.minZ - segment.a.Z) / delta.Z

	tHighX := (box.maxX - segment.a.X) / delta.X
	tHighY := (box.maxY - segment.a.Y) / delta.Y
	tHighZ := (box.maxZ - segment.a.Z) / delta.Z

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

func (box *compactAABB) intersectsAABB(other *CompactQueryAABB) bool {
	return (box.minX <= other.maxX) &&
		(box.maxX >= other.minX) &&
		(box.minY <= other.maxY) &&
		(box.maxY >= other.minY) &&
		(box.minZ <= other.maxZ) &&
		(box.maxZ >= other.minZ)
}

type compactCube struct {
	x float32
	y float32
	z float32
	r float32
}
