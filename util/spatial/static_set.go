package spatial

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

var staticSetLogger = spatialLogger.Path("/static-set")

// StaticSetSettings contains the settings for a StaticSet.
type StaticSetSettings struct {

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the set. This allows the set to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[int32]
}

// NewStaticSet creates a new StaticSet using the provided settings.
func NewStaticSet[T any](settings StaticSetSettings) *StaticSet[T] {
	initialItemCapacity := int32(1024)
	if settings.InitialItemCapacity.Specified {
		initialItemCapacity = settings.InitialItemCapacity.Value
		if initialItemCapacity < 0 {
			panic("initial item capacity must not be negative")
		}
	}
	return &StaticSet[T]{
		items: make([]staticSetItem[T], 0, initialItemCapacity),
	}
}

// StaticSet is a spatial structure that uses brute force to figure out which
// items are visible during a Visit.
//
// It allows only the insertion of static items. Such items cannot be resized,
// repositioned, or removed from the set.
type StaticSet[T any] struct {
	items []staticSetItem[T]
}

// Insert adds an item to this set.
func (t *StaticSet[T]) Insert(position dprec.Vec3, radius float64, item T) {
	if len(t.items) == cap(t.items) {
		staticSetLogger.Warn("Item slice capacity %d reached. Will grow.", len(t.items))
	}
	t.items = append(t.items, staticSetItem[T]{
		position: position,
		radius:   radius,
		item:     item,
	})
}

// VisitHexahedronRegion finds all items that are inside or intersect the
// specified hexahedron region. It calls the specified visitor for each item
// found.
func (t *StaticSet[T]) VisitHexahedronRegion(region *HexahedronRegion, visitor Visitor[T]) {
	for _, item := range t.items {
		if item.isInsideHexahedronRegion(region) {
			visitor.Visit(item.item)
		}
	}
}

type staticSetItem[T any] struct {
	position dprec.Vec3
	radius   float64
	item     T
}

func (i *staticSetItem[T]) isInsideHexahedronRegion(region *HexahedronRegion) bool {
	return region[0].ContainsSphere(i.position, i.radius) &&
		region[1].ContainsSphere(i.position, i.radius) &&
		region[2].ContainsSphere(i.position, i.radius) &&
		region[3].ContainsSphere(i.position, i.radius) &&
		region[4].ContainsSphere(i.position, i.radius) &&
		region[5].ContainsSphere(i.position, i.radius)
}
