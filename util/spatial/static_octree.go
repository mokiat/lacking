package spatial

import (
	"slices"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

var sizeToDoubleRadius = dprec.Sqrt(3)

const unspecifiedIndex = int32(-1)

// StaticOctreeStats represents the current state of a StaticOctree.
type StaticOctreeStats struct {
	NodeCount          int32
	ItemCount          int32
	ItemsCountPerDepth []int32
}

// StaticOctreeVisitStats represents statistics on the last visit operation
// performed on a StaticOctree.
type StaticOctreeVisitStats struct {
	NodeCount         int32
	NodeCountVisited  int32
	NodeCountAccepted int32
	NodeCountRejected int32
	ItemCount         int32
	ItemCountVisited  int32
	ItemCountAccepted int32
	ItemCountRejected int32
}

// StaticOctreeSettings contains the settings for a StaticOctree.
type StaticOctreeSettings struct {

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

// NewStaticOctree creates a new StaticOctree using the provided settings.
func NewStaticOctree[T any](settings StaticOctreeSettings) *StaticOctree[T] {
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

	nodes := make([]staticOctreeNode[T], 0, initialNodeCapacity)
	nodes = append(nodes, staticOctreeNode[T]{
		itemStart: unspecifiedIndex,
		itemEnd:   unspecifiedIndex,
		children: [8]int32{
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
			unspecifiedIndex, unspecifiedIndex, unspecifiedIndex, unspecifiedIndex,
		},
		x: 0.0,
		y: 0.0,
		z: 0.0,
		s: size,
	})
	items := make([]staticOctreeItem[T], 0, initialItemCapacity)

	return &StaticOctree[T]{
		nodes:     nodes,
		items:     items,
		biasRatio: biasRatio,
		maxDepth:  maxDepth,
	}
}

// StaticOctree represents an octree spatial structure that only allows the
// insertion of static items. Such items cannot be resized, repositioned, or
// removed from the tree.
type StaticOctree[T any] struct {
	nodes []staticOctreeNode[T]
	items []staticOctreeItem[T]

	biasRatio float64
	maxDepth  int32

	nodeCountAccepted int32
	nodeCountRejected int32
	itemCountAccepted int32
	itemCountRejected int32

	isDirty bool
}

// Stats returns statistics on the current state of this octree.
func (t *StaticOctree[T]) Stats() StaticOctreeStats {
	itemCountPerDepth := make([]int32, t.maxDepth+1)
	for i := int32(1); i <= t.maxDepth; i++ {
		itemCountPerDepth[i] = t.itemsAtDepth(0, 1, i)
	}
	return StaticOctreeStats{
		NodeCount:          int32(len(t.nodes)),
		ItemCount:          int32(len(t.items)),
		ItemsCountPerDepth: itemCountPerDepth,
	}
}

// VisitStats returns statistics information on the last executed search in
// this octree.
func (t *StaticOctree[T]) VisitStats() StaticOctreeVisitStats {
	return StaticOctreeVisitStats{
		NodeCount:         int32(len(t.nodes)),
		NodeCountVisited:  t.nodeCountAccepted + t.nodeCountRejected,
		NodeCountAccepted: t.nodeCountAccepted,
		NodeCountRejected: t.nodeCountRejected,
		ItemCount:         int32(len(t.items)),
		ItemCountVisited:  t.itemCountAccepted + t.itemCountRejected,
		ItemCountAccepted: t.itemCountAccepted,
		ItemCountRejected: t.itemCountRejected,
	}
}

// Insert adds an item to this octree at the specified position and taking the
// specified radius into account.
func (t *StaticOctree[T]) Insert(position dprec.Vec3, radius float64, item T) {
	if len(t.items) == cap(t.items) {
		logger.Warn("Item slice capacity (%d) reached for static octree! Will grow.", len(t.items))
	}
	t.isDirty = true
	nodeIndex := t.pickNodeForItem(position, radius)
	t.items = append(t.items, staticOctreeItem[T]{
		node:     nodeIndex,
		position: position,
		radius:   radius,
		item:     item,
	})
}

// VisitHexahedronRegion finds all items that are inside or intersect the
// specified hexahedron region. It calls the specified visitor for each item
// found.
func (t *StaticOctree[T]) VisitHexahedronRegion(region *HexahedronRegion, visitor Visitor[T]) {
	t.nodeCountAccepted = 0
	t.nodeCountRejected = 0
	t.itemCountAccepted = 0
	t.itemCountRejected = 0
	if t.isDirty {
		t.refresh()
	}
	t.visitNodeInHexahedronRegion(0, region, visitor)
}

func (t *StaticOctree[T]) pickNodeForItem(position dprec.Vec3, radius float64) int32 {
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

func (t *StaticOctree[T]) pickChildNode(parentNodeIndex int32, position dprec.Vec3, radius float64) int32 {
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
	childNodeIndex := int32(len(t.nodes)) // predict next node index
	parentNode.children[childIndex] = childNodeIndex

	if len(t.nodes) == cap(t.nodes) {
		logger.Warn("Node slice capacity (%d) reached for static octree! Will grow.", len(t.nodes))
	}
	// NOTE: DO NOT use parentNode after this append as the ref might be towards
	// an old slice.
	t.nodes = append(t.nodes, staticOctreeNode[T]{
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

func (t *StaticOctree[T]) visitNodeInHexahedronRegion(nodeIndex int32, region *HexahedronRegion, visitor Visitor[T]) {
	if nodeIndex == unspecifiedIndex {
		return
	}
	node := &t.nodes[nodeIndex]
	if node.isInsideHexahedronRegion(region) {
		t.nodeCountAccepted++
		for itemIndex := node.itemStart; itemIndex < node.itemEnd; itemIndex++ {
			item := &t.items[itemIndex]
			if item.isInsideHexahedronRegion(region) {
				visitor.Visit(item.item)
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

func (t *StaticOctree[T]) refresh() {
	// TODO: Test if this can be made faster with ref sorting.
	slices.SortFunc(t.items, compareStaticItems[T])

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

	t.isDirty = false
}

func (t *StaticOctree[T]) itemsAtDepth(nodeIndex, currentDepth, depth int32) int32 {
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

func compareStaticItems[T any](a, b staticOctreeItem[T]) int {
	return int(a.node - b.node)
}

type staticOctreeNode[T any] struct {
	itemStart int32
	itemEnd   int32
	children  [8]int32

	x float64
	y float64
	z float64
	s float64
}

func (n *staticOctreeNode[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	position := dprec.NewVec3(n.x, n.y, n.z)
	radius := n.s * sizeToDoubleRadius
	return region[0].ContainsSphere(position, radius) &&
		region[1].ContainsSphere(position, radius) &&
		region[2].ContainsSphere(position, radius) &&
		region[3].ContainsSphere(position, radius) &&
		region[4].ContainsSphere(position, radius) &&
		region[5].ContainsSphere(position, radius)
}

type staticOctreeItem[T any] struct {
	node     int32
	position dprec.Vec3
	radius   float64
	item     T
}

func (i *staticOctreeItem[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	return region[0].ContainsSphere(i.position, i.radius) &&
		region[1].ContainsSphere(i.position, i.radius) &&
		region[2].ContainsSphere(i.position, i.radius) &&
		region[3].ContainsSphere(i.position, i.radius) &&
		region[4].ContainsSphere(i.position, i.radius) &&
		region[5].ContainsSphere(i.position, i.radius)
}
