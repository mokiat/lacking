package query2d

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// BagItemID is an identifier used to control the placement of an item into a
// [Bag].
type BagItemID uint32

// nilItemIndex marks an id mapping whose item has been removed from the
// bag.
const nilItemIndex = int32(-1)

// BagSettings contains the settings for a [Bag].
type BagSettings struct {

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the bag. This allows the bag to preallocate memory
	// and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[uint32]
}

// Bag is a spatial structure that keeps its items in a flat list and answers
// queries by testing every item in turn.
//
// Unlike [Quadtree], a Bag performs no spatial subdivision, so the cost of a
// query grows linearly with the number of items. This makes it a good fit for
// small item counts, or for items that move so frequently that maintaining a
// hierarchy would not pay off. A Bag exposes the same insertion and query API
// as [Quadtree], so the two can be swapped with minimal changes to calling
// code.
type Bag[T any] struct {
	items       []bagItem[T]
	idMappings  []int32
	freeItemIDs *ds.Stack[BagItemID]
}

// NewBag creates a new [Bag] using the provided settings.
func NewBag[T any](settings BagSettings) *Bag[T] {
	initialItemCapacity := settings.InitialItemCapacity.ValueOrDefault(1024)
	return &Bag[T]{
		items:       make([]bagItem[T], 0, initialItemCapacity),
		idMappings:  make([]int32, 0, initialItemCapacity),
		freeItemIDs: ds.EmptyStack[BagItemID](),
	}
}

// Insert adds an item, which occupies the specified axis-aligned bounding box,
// to this bag.
//
// The box must not be empty (as per [shape2d.AABB.IsEmpty]), otherwise this
// function panics.
func (b *Bag[T]) Insert(aabb shape2d.AABB, value T) BagItemID {
	if aabb.IsEmpty() {
		panic("cannot insert item with empty area")
	}

	itemIndex := int32(len(b.items))
	if b.freeItemIDs.IsEmpty() {
		id := BagItemID(len(b.idMappings))
		b.idMappings = append(b.idMappings, itemIndex)
		b.items = append(b.items, bagItem[T]{
			id:        id,
			tightArea: newBoundingBoxFromAABB(aabb),
			value:     value,
		})
		return id
	}
	id := b.freeItemIDs.Pop()
	b.idMappings[id] = itemIndex
	b.items = append(b.items, bagItem[T]{
		id:        id,
		tightArea: newBoundingBoxFromAABB(aabb),
		value:     value,
	})
	return id
}

// Update repositions and resizes the item with the specified id to the new
// axis-aligned bounding box.
//
// As with [Bag.Insert], the box must not be empty, otherwise this function
// panics. Updating an item that has already been removed panics as well.
func (b *Bag[T]) Update(id BagItemID, aabb shape2d.AABB) {
	if aabb.IsEmpty() {
		panic("cannot update item to empty area")
	}

	itemIndex := b.idMappings[id]
	if itemIndex == nilItemIndex {
		panic("cannot update removed item")
	}
	b.items[itemIndex].tightArea = newBoundingBoxFromAABB(aabb)
}

// Remove removes the item with the specified id from this bag.
func (b *Bag[T]) Remove(id BagItemID) {
	itemIndex := b.idMappings[id]
	if itemIndex == nilItemIndex {
		panic("cannot remove item twice")
	}

	// Swap the removed item with the last one to keep the item slice tightly
	// packed, which is what makes the linear scan fast.
	lastIndex := int32(len(b.items) - 1)
	if itemIndex != lastIndex {
		moved := b.items[lastIndex]
		b.items[itemIndex] = moved
		b.idMappings[moved.id] = itemIndex
	}
	b.items[lastIndex] = gog.Zero[bagItem[T]]() // release references for garbage collection
	b.items = b.items[:lastIndex]

	b.idMappings[id] = nilItemIndex
	b.freeItemIDs.Push(id)
}

// QuerySegment finds all items that intersect the specified segment. Each
// found item is passed to the specified yield function. The order in which
// items are passed is undefined and might change between invocations.
func (b *Bag[T]) QuerySegment(segment shape2d.Segment, yield VisitorFunc[T]) {
	for i := range b.items {
		item := &b.items[i]
		if item.tightArea.intersectsSegment(&segment) {
			if !yield(item.value) {
				return
			}
		}
	}
}

// QueryAABB finds all items that are inside or intersect the specified
// axis-aligned bounding box. Each found item is passed to the specified yield
// function. The order in which items are passed is undefined and might change
// between invocations.
func (b *Bag[T]) QueryAABB(aabb shape2d.AABB, yield VisitorFunc[T]) {
	for i := range b.items {
		item := &b.items[i]
		if item.tightArea.intersectsAABB(&aabb) {
			if !yield(item.value) {
				return
			}
		}
	}
}

// QueryFrustum finds all items that are inside or intersect the specified
// frustum. Each found item is passed to the specified yield function. The
// order in which items are passed is undefined and might change between
// invocations.
//
// The test is conservative: an item is passed when its bounding box is not
// fully behind any of the four surfaces, so a box near an edge or a corner of
// the frustum may be passed even though it does not truly overlap it.
func (b *Bag[T]) QueryFrustum(frustum shape2d.Frustum, yield VisitorFunc[T]) {
	for i := range b.items {
		item := &b.items[i]
		if item.tightArea.intersectsFrustum(&frustum, allFrustumSurfaces) {
			if !yield(item.value) {
				return
			}
		}
	}
}

type bagItem[T any] struct {
	id        BagItemID
	tightArea boundingBox
	value     T
}
