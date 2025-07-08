package spatial

import (
	"log/slog"
	"slices"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// InvalidCubeOctreeItemID can be used to mark a reference as invalid.
const InvalidCubeOctreeItemID = CubeOctreeItemID(0xFFFFFFFF)

// CubeOctreeItemID is an identifier used to control the placement of an item
// into a dynamic octree.
type CubeOctreeItemID uint32

// CubeOctreeStats represents the current state of a CubeOctree.
type CubeOctreeStats struct {
	NodeCount          int32
	ItemCount          int32
	ItemsCountPerDepth []int32
}

// CubeOctreeVisitStats represents statistics on the last visit operation
// performed on a CubeOctree.
type CubeOctreeVisitStats struct {
	NodeCount         int32
	NodeCountVisited  int32
	NodeCountAccepted int32
	NodeCountRejected int32
	ItemCount         int32
	ItemCountVisited  int32
	ItemCountAccepted int32
	ItemCountRejected int32
}

// CubeOctreeSettings contains the settings for a CubeOctree.
type CubeOctreeSettings struct {

	// Size specifies the dimension of the octree cube. If not specified,
	// then a default size of 4096 will be used. Inserting an item outside
	// these bounds will be suboptimal (placed in root node).
	Size opt.T[float64]

	// MaxDepth controls the maximum depth that the octree could reach. If not
	// specified, then a default max depth of 8 will be used.
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

// NewCubeOctree creates a new CubeOctree using the provided settings.
func NewCubeOctree[T any](settings CubeOctreeSettings) *CubeOctree[T] {
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

	nodes := make([]cubeOctreeNode, 0, initialNodeCapacity)
	nodes = append(nodes, cubeOctreeNode{
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
			r: size, // using size, since loose cube has twice the radius
		},
	})
	items := make([]cubeOctreeItem[T], 0, initialItemCapacity)
	idMappings := make([]int32, 0, initialItemCapacity)

	return &CubeOctree[T]{
		nodes:           nodes,
		items:           items,
		freeNodeIndices: ds.NewStack[int32](32),
		freeItemIndices: ds.NewStack[int32](32),
		idMappings:      idMappings,
		maxDepth:        maxDepth,
	}
}

// CubeOctree is a spatial structure that uses a loose octree implementation
// with biased placement to enable the fast search of items within a region.
type CubeOctree[T any] struct {
	nodes           []cubeOctreeNode
	items           []cubeOctreeItem[T]
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
func (t *CubeOctree[T]) Stats() CubeOctreeStats {
	itemCountPerDepth := make([]int32, t.maxDepth+1)
	for i := int32(1); i <= t.maxDepth; i++ {
		itemCountPerDepth[i] = t.itemsAtDepth(0, 1, i)
	}
	return CubeOctreeStats{
		NodeCount:          int32(len(t.nodes)),
		ItemCount:          int32(len(t.items)),
		ItemsCountPerDepth: itemCountPerDepth,
	}
}

// VisitStats returns statistics information on the last executed search in
// this octree.
func (t *CubeOctree[T]) VisitStats() CubeOctreeVisitStats {
	return CubeOctreeVisitStats{
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
func (t *CubeOctree[T]) Insert(area CubeArea, value T) CubeOctreeItemID {
	t.isDirty = true
	if !t.freeItemIndices.IsEmpty() {
		itemIndex := t.freeItemIndices.Pop()
		item := &t.items[itemIndex]
		item.area = area
		item.value = value
		item.node = t.pickNodeForItem(area)
		return item.id
	} else {
		if len(t.items) == cap(t.items) {
			logger.Warn("Growing item capacity for cube octree.",
				slog.Int("capacity", len(t.items)),
			)
		}
		id := CubeOctreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		t.items = append(t.items, cubeOctreeItem[T]{
			id:    id,
			node:  t.pickNodeForItem(area),
			area:  area,
			value: value,
		})
		return id
	}
}

// Update repositions the item with the specified id to the new area.
func (t *CubeOctree[T]) Update(id CubeOctreeItemID, area CubeArea) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.area = area
	item.node = t.pickNodeForItem(area)
}

// Remove removes the item with the specified id from this data structure.
func (t *CubeOctree[T]) Remove(id CubeOctreeItemID) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.node = unspecifiedIndex
	t.freeItemIndices.Push(itemIndex)
}

// VisitArea finds all items that are inside or intersect the specified area.
// It calls the specified visitor for each item found.
func (t *CubeOctree[T]) VisitArea(area *CubeArea, visitor Visitor[T]) {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
	t.refresh()
	t.visitNodeInArea(0, area, visitor)
}

// GC runs cleanup and optimization logic. You should call that once
// per frame.
func (t *CubeOctree[T]) GC() {
	t.refresh()
}

func (t *CubeOctree[T]) pickNodeForItem(area CubeArea) int32 {
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

func (t *CubeOctree[T]) pickChildNode(parentNodeIndex int32, area CubeArea) int32 {
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
			logger.Warn("Will grow node slice capacity for cube octree.",
				slog.Int("capacity", len(t.nodes)),
			)
		}
		childNodeIndex := int32(len(t.nodes)) // predict next node index
		parentNode.children[childIndex] = childNodeIndex
		// NOTE: Do NOT use parentNode after this append as the ref might be towards
		// an old slice.
		t.nodes = append(t.nodes, cubeOctreeNode{
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

func (t *CubeOctree[T]) visitNodeInArea(nodeIndex int32, area *CubeArea, visitor Visitor[T]) {
	node := &t.nodes[nodeIndex]
	if area.Intersects(node.looseArea) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if area.Intersects(item.area) {
				visitor.Visit(item.value)
				t.itemCountAccepted++
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			if childNodeIndex != unspecifiedIndex {
				t.visitNodeInArea(childNodeIndex, area, visitor)
			}
		}
	} else {
		t.nodeCountRejected++
	}
}

func (t *CubeOctree[T]) refresh() {
	if t.isDirty {
		t.sortItems()
		t.eraseItemOffsets()
		t.evaluateItemOffsets()
		t.updateIDMappings()
		t.gcNodes()
		t.isDirty = false
	}
}

func (t *CubeOctree[T]) sortItems() {
	// TODO: Test if this can be made faster with ref sorting.
	slices.SortFunc(t.items, compareCubeOctreeItems[T])
}

func (t *CubeOctree[T]) eraseItemOffsets() {
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart = unspecifiedIndex
		node.itemEnd = unspecifiedIndex
	}
}

func (t *CubeOctree[T]) evaluateItemOffsets() {
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

func (t *CubeOctree[T]) updateIDMappings() {
	for i, item := range t.items {
		t.idMappings[item.id] = int32(i)
	}
}

func (t *CubeOctree[T]) gcNodes() {
	for i := range t.nodes {
		t.gcNode(int32(i))
	}
}

func (t *CubeOctree[T]) gcNode(nodeIndex int32) {
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

func (t *CubeOctree[T]) itemsAtDepth(nodeIndex, currentDepth, depth int32) int32 {
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

func compareCubeOctreeItems[T any](a, b cubeOctreeItem[T]) int {
	return int(a.node - b.node)
}

type cubeOctreeNode struct {
	parent    int32
	children  [8]int32
	looseArea CubeArea
	itemStart int32
	itemEnd   int32
}

func (n *cubeOctreeNode) isEmpty() bool {
	for _, childIndex := range n.children {
		if childIndex != unspecifiedIndex {
			return false
		}
	}
	return n.itemStart >= n.itemEnd
}

type cubeOctreeItem[T any] struct {
	id    CubeOctreeItemID
	node  int32
	area  CubeArea
	value T
}

// CubeAreaFromSphere creates a CubeArea that wraps a sphere.
func CubeAreaFromSphere(position dprec.Vec3, radius float64) CubeArea {
	return CubeArea{
		x: position.X,
		y: position.Y,
		z: position.Z,
		r: radius,
	}
}

// CubeArea represents an area in the shape of a cube.
type CubeArea struct {
	x float64
	y float64
	z float64
	r float64
}

// Intersects checks whether the area intersects another area.
func (a CubeArea) Intersects(other CubeArea) bool {
	dX := a.x - other.x
	dY := a.y - other.y
	dZ := a.z - other.z
	sR := a.r + other.r
	return (dX <= sR) && (dX >= -sR) && (dY <= sR) && (dY >= -sR) && (dZ <= sR) && (dZ >= -sR)
}
