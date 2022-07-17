package spatial

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/log"
)

var sizeToDoubleRadius = sprec.Sqrt(3)

// NewOctree creates a new Octree instance using the specified size and depth.
func NewOctree[T any](size float32, depth int) *Octree[T] {
	return &Octree[T]{
		depth: depth,
		root: &octreeNode[T]{
			size: size,
			head: &OctreeItem[T]{
				position: sprec.ZeroVec3(),
				radius:   size * sizeToDoubleRadius,
			},
		},
	}
}

// Octree represents an octree data structure that can be used to quickly
// find items within a given viewing area or custom region, without having
// to go through all available items.
//
// This particular implementation uses the loose octree approach.
type Octree[T any] struct {
	depth int
	root  *octreeNode[T]
}

// PrintDebug prints basic information that can be used for troubleshooting and
// optimization.
func (t *Octree[T]) PrintDebug() {
	log.Info("------- OCTREE -------")
	for i := 1; i <= t.depth; i++ {
		log.Info("Items at depth %-2d: %d", i, t.itemsAtDepth(t.root, 1, i))
	}
	log.Info("----------------------")
}

// CreateItem creates and positions a new item in this Octree.
func (t *Octree[T]) CreateItem(value T) *OctreeItem[T] {
	item := &OctreeItem[T]{
		tree:     t,
		position: sprec.ZeroVec3(),
		radius:   1.0,
		value:    value,
	}
	item.invalidate()
	return item
}

func (t *Octree[T]) VisitHexahedronRegion(region *HexahedronRegion, cb func(item T)) {
	t.visitNodeInHexahedronRegion(t.root, region, cb)
}

func (t *Octree[T]) visitNodeInHexahedronRegion(node *octreeNode[T], region *HexahedronRegion, cb func(item T)) {
	if node == nil {
		return
	}
	if !node.head.isInsideHexahedronRegion(region) {
		return
	}
	for item := node.head.next; item != nil; item = item.next {
		if item.isInsideHexahedronRegion(region) {
			cb(item.value)
		}
	}
	for i := 0; i < 8; i++ {
		t.visitNodeInHexahedronRegion(node.children[i], region, cb)
	}
}

func (t *Octree[T]) add(item *OctreeItem[T]) {
	bestNode := t.root
	depth := 0
	for node := bestNode; node != nil; node = t.pickChildNode(node, item, depth) {
		bestNode = node
		depth++
	}
	item.next = bestNode.head.next
	item.prev = bestNode.head
	if bestNode.head.next != nil {
		bestNode.head.next.prev = item
	}
	bestNode.head.next = item
}

func (t *Octree[T]) remove(item *OctreeItem[T]) {
	if item.prev != nil {
		item.prev.next = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	}
	item.prev = nil
	item.next = nil
}

func (t *Octree[T]) pickChildNode(parent *octreeNode[T], item *OctreeItem[T], depth int) *octreeNode[T] {
	if depth >= t.depth {
		return nil // there are no children nodes
	}

	childSize := parent.size / 2.0
	childHalfSize := childSize / 2.0
	if item.radius > childHalfSize {
		return nil // no child will be able to fit this
	}

	// it has to be one of the eight children
	distanceManhattan := sprec.Vec3Diff(item.position, parent.head.position)
	var (
		childIndex    = 0
		childPosition = parent.head.position
	)
	if distanceManhattan.X < 0.0 {
		childPosition.X -= childHalfSize
	} else {
		childIndex += 1
		childPosition.X += childHalfSize
	}
	if distanceManhattan.Z < 0.0 {
		childPosition.Z -= childHalfSize
	} else {
		childIndex += 2
		childPosition.Z += childHalfSize
	}
	if distanceManhattan.Y < 0.0 {
		childIndex += 4
		childPosition.Y -= childHalfSize
	} else {
		childPosition.Y += childHalfSize
	}

	if parent.children[childIndex] != nil {
		return parent.children[childIndex]
	}

	childNode := &octreeNode[T]{
		size: childSize,
		head: &OctreeItem[T]{
			position: childPosition,
			radius:   childSize * sizeToDoubleRadius,
		},
	}
	parent.children[childIndex] = childNode
	return childNode
}

func (t *Octree[T]) itemsAtDepth(node *octreeNode[T], currentDepth, depth int) int {
	if currentDepth == depth {
		return node.itemCount()
	}
	result := 0
	for i := 0; i < 8; i++ {
		if child := node.children[i]; child != nil {
			result += t.itemsAtDepth(child, currentDepth+1, depth)
		}
	}
	return result
}

// OctreeItem represents an item that can be placed inside an Octree.
type OctreeItem[T any] struct {
	tree *Octree[T]
	prev *OctreeItem[T]
	next *OctreeItem[T]

	position sprec.Vec3
	radius   float32
	value    T
}

// Delete removes this item from its Octree.
func (i *OctreeItem[T]) Delete() {
	// TODO: Cache items in a pool
	i.tree.remove(i)
	i.tree = nil
}

// Position returns the world position of this item.
func (i *OctreeItem[T]) Position() sprec.Vec3 {
	return i.position
}

// SetPosition changes the world position of this item to the specified value.
func (i *OctreeItem[T]) SetPosition(position sprec.Vec3) {
	if position != i.position {
		i.position = position
		i.invalidate()
	}
}

// Radius returns the bounding sphere radius of this item. This is used to
// determine visibility of the item.
func (i *OctreeItem[T]) Radius() float32 {
	return i.radius
}

// SetRadius changes the bounding sphere radius of this item.
func (i *OctreeItem[T]) SetRadius(radius float32) {
	if radius != i.radius {
		i.radius = radius
		i.invalidate()
	}
}

func (i *OctreeItem[T]) invalidate() {
	// TODO: Mark as dirty instead and relocate only once a Visit is performed.
	i.tree.remove(i)
	i.tree.add(i)
}

func (i *OctreeItem[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	return region[0].ContainsSphere(i.position, i.radius) &&
		region[1].ContainsSphere(i.position, i.radius) &&
		region[2].ContainsSphere(i.position, i.radius) &&
		region[3].ContainsSphere(i.position, i.radius) &&
		region[4].ContainsSphere(i.position, i.radius) &&
		region[5].ContainsSphere(i.position, i.radius)
}

type octreeNode[T any] struct {
	size     float32
	children [8]*octreeNode[T]
	head     *OctreeItem[T]
}

func (n *octreeNode[T]) itemCount() int {
	result := 0
	for item := n.head.next; item != nil; item = item.next {
		result++
	}
	return result
}
