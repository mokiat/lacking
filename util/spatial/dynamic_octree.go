package spatial

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"golang.org/x/exp/slices"
)

// DynamicOctreeItemID is an identifier used to control the placement of an item
// into a dynamic octree.
type DynamicOctreeItemID uint32

// DynamicOctreeStats represents the current state of a DynamicOctree.
type DynamicOctreeStats struct {
	NodeCount          int32
	ItemCount          int32
	ItemsCountPerDepth []int32
}

// DynamicOctreeVisitStats represents statistics on the last visit operation
// performed on a DynamicOctree.
type DynamicOctreeVisitStats struct {
	NodeCount         int32
	NodeCountVisited  int32
	NodeCountAccepted int32
	NodeCountRejected int32
	ItemCount         int32
	ItemCountVisited  int32
	ItemCountAccepted int32
	ItemCountRejected int32
}

// DynamicOctreeSettings contains the settings for a DynamicOctree.
type DynamicOctreeSettings struct {

	// Size specifies the dimension of the octree cube. If not specified,
	// then a default size of 4096 will be used. Inserting an item outside
	// these bounds will be suboptimal (placed in root node).
	Size opt.T[float64]

	// MaxDepth controls the maximum depth that the octree could reach. If not
	// specified, then a default max depth of 8 will be used.
	MaxDepth opt.T[int32]

	// BiasRatio is multiplied to the item radius in order to force items to
	// be placed upper in the octree hierarchy. This can lead to better
	// performance in certain cases, as it can prevent double visibility check
	// on both the item and the (tightly fit) node that contains it.
	//
	// The value must not be smaller than 1.0 and is 2.0 by default.
	BiasRatio opt.T[float64]

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

// NewDynamicOctree creates a new DynamicOctree using the provided settings.
func NewDynamicOctree[T any](settings DynamicOctreeSettings) *DynamicOctree[T] {
	size := 4096.0
	if settings.Size.Specified {
		size = settings.Size.Value
		if size < 1.0 {
			panic("size cannot be smaller than 1")
		}
	}
	maxDepth := int32(8)
	if settings.MaxDepth.Specified {
		maxDepth = settings.MaxDepth.Value
		if maxDepth < 1 {
			panic("max depth cannot be smaller than 1")
		}
	}
	biasRatio := 2.0
	if settings.BiasRatio.Specified {
		biasRatio = settings.BiasRatio.Value
		if biasRatio < 1.0 {
			panic("bias ratio cannot be smaller than 1")
		}
	}
	initialNodeCapacity := int32(4096)
	if settings.InitialNodeCapacity.Specified {
		initialNodeCapacity = settings.InitialNodeCapacity.Value
		if initialNodeCapacity < 0 {
			panic("initial node capacity must not be negative")
		}
	}
	initialItemCapacity := int32(1024)
	if settings.InitialItemCapacity.Specified {
		initialItemCapacity = settings.InitialItemCapacity.Value
		if initialItemCapacity < 0 {
			panic("initial item capacity must not be negative")
		}
	}

	nodes := make([]dynamicOctreeNode[T], 0, initialNodeCapacity)
	nodes = append(nodes, dynamicOctreeNode[T]{
		parent: unspecifiedIndex,
		children: [8]int32{
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
		},
		itemStart: unspecifiedIndex,
		itemEnd:   unspecifiedIndex,
		x:         0.0,
		y:         0.0,
		z:         0.0,
		s:         size,
	})
	items := make([]dynamicOctreeItem[T], 0, initialItemCapacity)
	idMappings := make([]int32, 0, initialItemCapacity)

	return &DynamicOctree[T]{
		nodes:           nodes,
		items:           items,
		freeNodeIndices: ds.NewStack[int32](32),
		freeItemIndices: ds.NewStack[int32](32),
		idMappings:      idMappings,
		biasRatio:       biasRatio,
		maxDepth:        maxDepth,
	}
}

// DynamicOctree is a spatial structure that uses a loose octree implementation
// with biased placement to enable the fast search of items within a region.
type DynamicOctree[T any] struct {
	nodes           []dynamicOctreeNode[T]
	items           []dynamicOctreeItem[T]
	freeNodeIndices *ds.Stack[int32]
	freeItemIndices *ds.Stack[int32]
	idMappings      []int32

	biasRatio float64
	maxDepth  int32

	nodeCountAccepted int32
	nodeCountRejected int32
	itemCountAccepted int32
	itemCountRejected int32

	isDirty bool
}

// Stats returns statistics on the current state of this octree.
func (t *DynamicOctree[T]) Stats() DynamicOctreeStats {
	itemCountPerDepth := make([]int32, t.maxDepth+1)
	for i := int32(1); i <= t.maxDepth; i++ {
		itemCountPerDepth[i] = t.itemsAtDepth(0, 1, i)
	}
	return DynamicOctreeStats{
		NodeCount:          int32(len(t.nodes)),
		ItemCount:          int32(len(t.items)),
		ItemsCountPerDepth: itemCountPerDepth,
	}
}

// VisitStats returns statistics information on the last executed search in
// this octree.
func (t *DynamicOctree[T]) VisitStats() DynamicOctreeVisitStats {
	return DynamicOctreeVisitStats{
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
func (t *DynamicOctree[T]) Insert(position dprec.Vec3, radius float64, value T) DynamicOctreeItemID {
	t.isDirty = true
	if !t.freeItemIndices.IsEmpty() {
		itemIndex := t.freeItemIndices.Pop()
		item := &t.items[itemIndex]
		item.position = position
		item.radius = radius
		item.value = value
		item.node = t.pickNodeForItem(position, radius)
		return item.id
	} else {
		if len(t.items) == cap(t.items) {
			logger.Warn("Item slice capacity (%d) reached for dynamic octree! Will grow.", len(t.items))
		}
		id := DynamicOctreeItemID(len(t.items))
		t.idMappings = append(t.idMappings, int32(id))
		nodeIndex := t.pickNodeForItem(position, radius)
		t.items = append(t.items, dynamicOctreeItem[T]{
			id:       id,
			node:     nodeIndex,
			position: position,
			radius:   radius,
			value:    value,
		})
		return id
	}
}

// Update repositions the item with the specified id to the new position and
// radius.
func (t *DynamicOctree[T]) Update(id DynamicOctreeItemID, position dprec.Vec3, radius float64) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.position = position
	item.radius = radius
	item.node = t.pickNodeForItem(position, radius)
}

// Remove removes the item with the specified id from this data structure.
func (t *DynamicOctree[T]) Remove(id DynamicOctreeItemID) {
	t.isDirty = true
	itemIndex := t.idMappings[id]
	item := &t.items[itemIndex]
	item.node = unspecifiedIndex
	t.freeItemIndices.Push(itemIndex)
}

// VisitHexahedronRegion finds all items that are inside or intersect the
// specified hexahedron region. It calls the specified visitor for each item
// found.
func (t *DynamicOctree[T]) VisitHexahedronRegion(region *HexahedronRegion, visitor Visitor[T]) {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
	t.refresh()
	t.visitNodeInHexahedronRegion(0, region, visitor)
}

func (t *DynamicOctree[T]) pickNodeForItem(position dprec.Vec3, radius float64) int32 {
	bestNodeIndex := unspecifiedIndex
	currentNodeIndex := int32(0)
	depth := int32(0)
	for currentNodeIndex != unspecifiedIndex {
		bestNodeIndex = currentNodeIndex
		depth++
		if depth >= t.maxDepth {
			break
		}
		currentNodeIndex = t.pickChildNode(currentNodeIndex, position, radius)
	}
	return bestNodeIndex
}

func (t *DynamicOctree[T]) pickChildNode(parentNodeIndex int32, position dprec.Vec3, radius float64) int32 {
	parentNode := &t.nodes[parentNodeIndex]
	childSize := parentNode.s / 2.0
	if radius*t.biasRatio > childSize {
		return unspecifiedIndex
	}
	childHalfSize := childSize / 2.0

	// It has to be inside one of the eight children.
	var (
		childIndex = 0
		childX     = parentNode.x
		childY     = parentNode.y
		childZ     = parentNode.z
	)
	if position.X < parentNode.x {
		childX -= childHalfSize
	} else {
		childIndex += 1
		childX += childHalfSize
	}
	if position.Z < parentNode.z {
		childZ -= childHalfSize
	} else {
		childIndex += 2
		childZ += childHalfSize
	}
	if position.Y < parentNode.y {
		childY -= childHalfSize
	} else {
		childIndex += 4
		childY += childHalfSize
	}

	if parentNode.children[childIndex] != unspecifiedIndex {
		return parentNode.children[childIndex]
	}

	if !t.freeNodeIndices.IsEmpty() {
		childNodeIndex := t.freeNodeIndices.Pop()
		parentNode.children[childIndex] = childNodeIndex
		childNode := &t.nodes[childNodeIndex]
		childNode.parent = parentNodeIndex
		childNode.x = childX
		childNode.y = childY
		childNode.z = childZ
		childNode.s = childSize
		return childNodeIndex
	} else {
		if len(t.nodes) == cap(t.nodes) {
			logger.Warn("Node slice capacity (%d) reached for dynamic octree! Will grow.", len(t.nodes))
		}
		childNodeIndex := int32(len(t.nodes)) // predict next node index
		parentNode.children[childIndex] = childNodeIndex
		// NOTE: DO NOT use parentNode after this append as the ref might be towards
		// an old slice.
		t.nodes = append(t.nodes, dynamicOctreeNode[T]{
			parent:    parentNodeIndex,
			itemStart: unspecifiedIndex,
			itemEnd:   unspecifiedIndex,
			children: [8]int32{
				unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
				unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
			},
			x: childX,
			y: childY,
			z: childZ,
			s: childSize,
		})
		return childNodeIndex
	}
}

func (t *DynamicOctree[T]) visitNodeInHexahedronRegion(nodeIndex int32, region *HexahedronRegion, visitor Visitor[T]) {
	if nodeIndex == unspecifiedIndex {
		return
	}
	node := &t.nodes[nodeIndex]
	if node.isInsideHexahedronRegion(region) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if item.isInsideHexahedronRegion(region) {
				visitor.Visit(item.value)
				t.itemCountAccepted++
			} else {
				t.itemCountRejected++
			}
		}
		for _, childNodeIndex := range node.children {
			t.visitNodeInHexahedronRegion(childNodeIndex, region, visitor)
		}
	} else {
		t.nodeCountRejected++
	}
}

func (t *DynamicOctree[T]) refresh() {
	if t.isDirty {
		t.sortItems()
		t.eraseItemOffsets()
		t.evaluateItemOffsets()
		t.updateIDMappings()
		t.gcNodes()
		t.isDirty = false
	}
}

func (t *DynamicOctree[T]) sortItems() {
	slices.SortFunc(t.items, compareDynamicOctreeItems[T])
}

func (t *DynamicOctree[T]) eraseItemOffsets() {
	for i := range t.nodes {
		node := &t.nodes[i]
		node.itemStart = unspecifiedIndex
		node.itemEnd = unspecifiedIndex
	}
}

func (t *DynamicOctree[T]) evaluateItemOffsets() {
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

func (t *DynamicOctree[T]) updateIDMappings() {
	for i, item := range t.items {
		t.idMappings[item.id] = int32(i)
	}
}

func (t *DynamicOctree[T]) gcNodes() {
	for i := range t.nodes {
		t.gcNode(int32(i))
	}
}

func (t *DynamicOctree[T]) gcNode(nodeIndex int32) {
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

func (t *DynamicOctree[T]) itemsAtDepth(nodeIndex, currentDepth, depth int32) int32 {
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

func compareDynamicOctreeItems[T any](a, b dynamicOctreeItem[T]) int {
	return int(a.node - b.node)
}

type dynamicOctreeNode[T any] struct {
	parent    int32
	children  [8]int32
	itemStart int32
	itemEnd   int32

	x float64
	y float64
	z float64
	s float64
}

func (n *dynamicOctreeNode[T]) isEmpty() bool {
	for _, childIndex := range n.children {
		if childIndex != unspecifiedIndex {
			return false
		}
	}
	return n.itemStart >= n.itemEnd
}

func (n *dynamicOctreeNode[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	position := dprec.NewVec3(n.x, n.y, n.z)
	radius := n.s * sizeToDoubleRadius
	return region[0].ContainsSphere(position, radius) &&
		region[1].ContainsSphere(position, radius) &&
		region[2].ContainsSphere(position, radius) &&
		region[3].ContainsSphere(position, radius) &&
		region[4].ContainsSphere(position, radius) &&
		region[5].ContainsSphere(position, radius)
}

type dynamicOctreeItem[T any] struct {
	id       DynamicOctreeItemID
	node     int32
	position dprec.Vec3
	radius   float64
	value    T
}

func (i *dynamicOctreeItem[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	return region[0].ContainsSphere(i.position, i.radius) &&
		region[1].ContainsSphere(i.position, i.radius) &&
		region[2].ContainsSphere(i.position, i.radius) &&
		region[3].ContainsSphere(i.position, i.radius) &&
		region[4].ContainsSphere(i.position, i.radius) &&
		region[5].ContainsSphere(i.position, i.radius)
}
