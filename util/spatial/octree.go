package spatial

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/util/datastruct"
)

var sizeToDoubleRadius = dprec.Sqrt(3)

// Visitor represents a callback mechanism to pass items back to the client.
type Visitor[T any] interface {
	// Reset indicates that a new batch of items will be provided.
	Reset()
	// Visit is called for each observed item.
	Visit(item T)
}

// VisitorFunc is an implementation of Visitor that passes each observed
// item to the wrapped function.
type VisitorFunc[T any] func(item T)

// Reset does nothing and is just so Visitor interface is implemented.
func (f VisitorFunc[T]) Reset() {}

// Visit calls the wrapped function.
func (f VisitorFunc[T]) Visit(item T) {
	f(item)
}

// NewVisitorBucket creates a new NewVisitorBucket instance with the specified
// initial capacity, which is only used to preallocate memory. It is allowed
// to exceed the initial capacity.
func NewVisitorBucket[T any](initCapacity int) *VisitorBucket[T] {
	return &VisitorBucket[T]{
		items: make([]T, 0, initCapacity),
	}
}

// VisitorBucket is an implementation of Visitor that stores observed items
// into a buffer for faster and more cache-friendly iteration afterwards.
type VisitorBucket[T any] struct {
	items []T
}

// Reset rewinds the item buffer.
func (r *VisitorBucket[T]) Reset() {
	r.items = r.items[:0]
}

// Visit records the passed item into the buffer.
func (r *VisitorBucket[T]) Visit(item T) {
	r.items = append(r.items, item)
}

// Each calls the provided closure function for each item in the buffer.
func (r *VisitorBucket[T]) Each(cb func(item T)) {
	for _, item := range r.items {
		cb(item)
	}
}

// Items returns the items stored in the buffer. The returned slice is valid
// only until the Reset function is called.
func (r *VisitorBucket[T]) Items() []T {
	return r.items
}

// NewOctree creates a new Octree instance using the specified size and depth.
func NewOctree[T any](size float64, depth, capacity int) *Octree[T] {
	var (
		nodePool datastruct.Pool[octreeNode[T]]
		itemPool datastruct.Pool[OctreeItem[T]]
	)
	if capacity > 0 {
		nodePool = datastruct.NewStaticPool[octreeNode[T]](capacity)
		itemPool = datastruct.NewStaticPool[OctreeItem[T]](capacity)
	} else {
		nodePool = datastruct.NewDynamicPool[octreeNode[T]]()
		itemPool = datastruct.NewDynamicPool[OctreeItem[T]]()
	}
	root := nodePool.Fetch()
	*root = octreeNode[T]{
		head: &OctreeItem[T]{
			x:      0,
			y:      0,
			z:      0,
			radius: sizeToRadius(int32(size)),
		},
	}
	return &Octree[T]{
		size:     int32(size),
		depth:    int32(depth),
		nodePool: nodePool,
		itemPool: itemPool,
		root:     root,
	}
}

// Octree represents an octree data structure that can be used to quickly
// find items within a given viewing area or custom region, without having
// to go through all available items.
//
// This particular implementation uses the loose octree approach.
type Octree[T any] struct {
	size     int32
	depth    int32
	nodePool datastruct.Pool[octreeNode[T]]
	itemPool datastruct.Pool[OctreeItem[T]]
	root     *octreeNode[T]
}

// PrintDebug prints basic information that can be used for troubleshooting and
// optimization.
func (t *Octree[T]) PrintDebug() {
	log.Info("------- OCTREE -------")
	for i := int32(1); i <= t.depth; i++ {
		log.Info("Items at depth %-2d: %d", i, t.itemsAtDepth(t.root, 1, i))
	}
	log.Info("----------------------")
}

// CreateItem creates and positions a new item in this Octree.
func (t *Octree[T]) CreateItem(value T) *OctreeItem[T] {
	item := t.itemPool.Fetch()
	*item = OctreeItem[T]{
		tree:   t,
		x:      0,
		y:      0,
		z:      0,
		radius: 1,
		value:  value,
	}
	item.invalidate()
	return item
}

// VisitHexahedronRegion finds all items that are inside or intersect the
// specified hexahedron region. It calls the specified visitor for each found
// item.
func (t *Octree[T]) VisitHexahedronRegion(region *HexahedronRegion, visitor Visitor[T]) {
	visitor.Reset()
	t.visitNodeInHexahedronRegion(t.root, region, visitor)
}

func (t *Octree[T]) visitNodeInHexahedronRegion(node *octreeNode[T], region *HexahedronRegion, visitor Visitor[T]) {
	if node == nil {
		return
	}
	if !node.head.isInsideHexahedronRegion(region) {
		return
	}
	for item := node.head.next; item != nil; item = item.next {
		if item.isInsideHexahedronRegion(region) {
			visitor.Visit(item.value)
		}
	}
	for i := 0; i < 8; i++ {
		t.visitNodeInHexahedronRegion(node.children[i], region, visitor)
	}
}

func (t *Octree[T]) add(item *OctreeItem[T]) {
	bestNode := t.root
	depth := int32(0)
	parentSize := t.size * 2
	for node := bestNode; node != nil; node = t.pickChildNode(node, parentSize, item, depth) {
		bestNode = node
		parentSize /= 2
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

func (t *Octree[T]) pickChildNode(parent *octreeNode[T], parentSize int32, item *OctreeItem[T], depth int32) *octreeNode[T] {
	if depth >= t.depth || parentSize <= 2 {
		return nil // there are no children nodes
	}

	childSize := parentSize / 2
	childHalfSize := childSize / 2
	if item.radius > childHalfSize {
		return nil // no child will be able to fit this
	}

	// it has to be one of the eight children
	var (
		childIndex = 0
		childX     = parent.head.x
		childY     = parent.head.y
		childZ     = parent.head.z
	)
	if item.x < parent.head.x {
		childX -= childHalfSize
	} else {
		childIndex += 1
		childX += childHalfSize
	}
	if item.z < parent.head.z {
		childZ -= childHalfSize
	} else {
		childIndex += 2
		childZ += childHalfSize
	}
	if item.y < parent.head.y {
		childIndex += 4
		childY -= childHalfSize
	} else {
		childY += childHalfSize
	}

	if parent.children[childIndex] != nil {
		return parent.children[childIndex]
	}

	childNode := t.nodePool.Fetch()
	*childNode = octreeNode[T]{
		head: &OctreeItem[T]{
			x:      childX,
			y:      childY,
			z:      childZ,
			radius: sizeToRadius(childSize),
		},
	}
	parent.children[childIndex] = childNode
	return childNode
}

func (t *Octree[T]) itemsAtDepth(node *octreeNode[T], currentDepth, depth int32) int {
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

	x      int32
	y      int32
	z      int32
	radius int32
	value  T
}

// Delete removes this item from its Octree.
func (i *OctreeItem[T]) Delete() {
	i.tree.remove(i)
	i.tree.itemPool.Restore(i)
	i.tree = nil
}

// Position returns the world position of this item.
func (i *OctreeItem[T]) Position() dprec.Vec3 {
	return dprec.NewVec3(float64(i.x), float64(i.y), float64(i.z))
}

// SetPosition changes the world position of this item to the specified value.
func (i *OctreeItem[T]) SetPosition(position dprec.Vec3) {
	i.x = int32(position.X)
	i.y = int32(position.Y)
	i.z = int32(position.Z)
	i.invalidate()
}

// Radius returns the bounding sphere radius of this item. This is used to
// determine visibility of the item.
func (i *OctreeItem[T]) Radius() float64 {
	return float64(i.radius)
}

// SetRadius changes the bounding sphere radius of this item.
func (i *OctreeItem[T]) SetRadius(radius float64) {
	i.radius = int32(math.Ceil(radius))
	i.invalidate()
}

func (i *OctreeItem[T]) invalidate() {
	// IDEA: Mark as dirty instead and relocate only once a Visit is performed.
	i.tree.remove(i)
	i.tree.add(i)
}

func (i *OctreeItem[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	position, radius := i.Position(), i.Radius()
	return region[0].ContainsSphere(position, radius) &&
		region[1].ContainsSphere(position, radius) &&
		region[2].ContainsSphere(position, radius) &&
		region[3].ContainsSphere(position, radius) &&
		region[4].ContainsSphere(position, radius) &&
		region[5].ContainsSphere(position, radius)
}

type octreeNode[T any] struct {
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

func sizeToRadius(size int32) int32 {
	return int32(float64(size) * sizeToDoubleRadius)
}
