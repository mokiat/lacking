package spatial

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// DynamicSetSettings contains the settings for a DynamicSet.
type DynamicSetSettings struct {

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the set. This allows the set to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[int32]
}

// DynamicSetItemID is an identifier used to control the placement of an item
// into a dynamic set.
type DynamicSetItemID uint32

// NewDynamicSet creates a new DynamicSet using the provided settings.
func NewDynamicSet[T any](settings DynamicSetSettings) *DynamicSet[T] {
	initialItemCapacity := int32(1024)
	if settings.InitialItemCapacity.Specified {
		initialItemCapacity = settings.InitialItemCapacity.Value
		if initialItemCapacity < 0 {
			panic("initial item capacity must not be negative")
		}
	}
	return &DynamicSet[T]{
		items: make(map[DynamicSetItemID]dynamicSetItem[T], initialItemCapacity),
	}
}

// DynamicSet is a spatial structure that uses brute force to figure out which
// items are visible during a Visit.
//
// It allows only the insertion of static items. Such items cannot be resized,
// repositioned, or removed from the set.
type DynamicSet[T any] struct {
	items      map[DynamicSetItemID]dynamicSetItem[T]
	nextFreeID uint32
}

// Insert adds an item to this set.
func (t *DynamicSet[T]) Insert(position dprec.Vec3, radius float64, item T) DynamicSetItemID {
	t.nextFreeID++
	t.items[DynamicSetItemID(t.nextFreeID)] = dynamicSetItem[T]{
		position: position,
		radius:   radius,
		item:     item,
	}
	return DynamicSetItemID(t.nextFreeID)
}

// Update repositions the item with the specified id to the new
// position and radius.
func (t *DynamicSet[T]) Update(id DynamicSetItemID, position dprec.Vec3, radius float64) {
	item := t.items[id]
	item.position = position
	item.radius = radius
	t.items[id] = item
}

// Remove removes the item with the specified it from this data structure.
func (t *DynamicSet[T]) Remove(id DynamicSetItemID) {
	delete(t.items, id)
}

// VisitHexahedronRegion finds all items that are inside or intersect the
// specified hexahedron region. It calls the specified visitor for each item
// found.
func (t *DynamicSet[T]) VisitHexahedronRegion(region *HexahedronRegion, visitor Visitor[T]) {
	for _, item := range t.items {
		if item.isInsideHexahedronRegion(region) {
			visitor.Visit(item.item)
		}
	}
}

type dynamicSetItem[T any] struct {
	position dprec.Vec3
	radius   float64
	item     T
}

func (i *dynamicSetItem[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	return region[0].ContainsSphere(i.position, i.radius) &&
		region[1].ContainsSphere(i.position, i.radius) &&
		region[2].ContainsSphere(i.position, i.radius) &&
		region[3].ContainsSphere(i.position, i.radius) &&
		region[4].ContainsSphere(i.position, i.radius) &&
		region[5].ContainsSphere(i.position, i.radius)
}
