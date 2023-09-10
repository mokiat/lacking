package spatial

import (
	"math"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// SweepPruneItemID is an identifier used to control the placement of an item
// into the sweep and prune set.
type SweepPruneItemID uint32

// SweepPruneSetSettings contains the settings for a SweepPruneSet.
type SweepPruneSetSettings struct {

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the set. This allows the set to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[int32]
}

// NewSweepPruneSet creates a new SweepPruneSet using the provided settings.
func NewSweepPruneSet[T any](settings SweepPruneSetSettings) *SweepPruneSet[T] {
	initialItemCapacity := int32(1024)
	if settings.InitialItemCapacity.Specified {
		initialItemCapacity = settings.InitialItemCapacity.Value
		if initialItemCapacity < 0 {
			panic("initial item capacity must not be negative")
		}
	}
	return &SweepPruneSet[T]{
		items:        make([]sweepPruneItem[T], 0, initialItemCapacity),
		freeItemIDs:  ds.NewStack[uint32](32),
		dirtyItemIDs: make(map[uint32]struct{}, initialItemCapacity),
		seriesX:      make([]sweepPruneMarker, 0, initialItemCapacity*2),
		seriesY:      make([]sweepPruneMarker, 0, initialItemCapacity*2),
		seriesZ:      make([]sweepPruneMarker, 0, initialItemCapacity*2),
		candidates:   make(map[sweepPrunePair]uint8, 32*32),
		active:       make(map[uint32]struct{}, 32),
	}
}

// SweepPruneSet is a spatial data structure that uses the sweep and prune
// algorithm to determine potentially intersecting item pairs.
type SweepPruneSet[T any] struct {
	items        []sweepPruneItem[T]
	freeItemIDs  *ds.Stack[uint32]
	dirtyItemIDs map[uint32]struct{}
	seriesX      []sweepPruneMarker
	seriesY      []sweepPruneMarker
	seriesZ      []sweepPruneMarker
	candidates   map[sweepPrunePair]uint8
	active       map[uint32]struct{}
}

// Insert adds an item to this set.
func (s *SweepPruneSet[T]) Insert(position dprec.Vec3, radius float64, value T) SweepPruneItemID {
	var itemIndex uint32
	if !s.freeItemIDs.IsEmpty() {
		itemIndex = s.freeItemIDs.Pop()
	} else {
		itemIndex = uint32(len(s.items))
		s.items = append(s.items, sweepPruneItem[T]{})
		s.seriesX = append(s.seriesX, sweepPruneMarker{
			ItemIndex: itemIndex,
			IsStart:   true,
		})
		s.seriesX = append(s.seriesX, sweepPruneMarker{
			ItemIndex: itemIndex,
			IsStart:   false,
		})
		s.seriesY = append(s.seriesY, sweepPruneMarker{
			ItemIndex: itemIndex,
			IsStart:   true,
		})
		s.seriesY = append(s.seriesY, sweepPruneMarker{
			ItemIndex: itemIndex,
			IsStart:   false,
		})
		s.seriesZ = append(s.seriesZ, sweepPruneMarker{
			ItemIndex: itemIndex,
			IsStart:   true,
		})
		s.seriesZ = append(s.seriesZ, sweepPruneMarker{
			ItemIndex: itemIndex,
			IsStart:   false,
		})
	}
	item := &s.items[itemIndex]
	item.Position = position
	item.Radius = radius
	item.Value = value
	s.dirtyItemIDs[itemIndex] = struct{}{}
	return SweepPruneItemID(itemIndex)
}

// Update repositions the item with the specified id to the new position
// and radius.
func (s *SweepPruneSet[T]) Update(id SweepPruneItemID, position dprec.Vec3, radius float64) {
	itemIndex := uint32(id)
	item := &s.items[itemIndex]
	item.Position = position
	item.Radius = radius
	s.dirtyItemIDs[itemIndex] = struct{}{}
}

// Remove removes the item with the specified id from this data structure.
func (s *SweepPruneSet[T]) Remove(id SweepPruneItemID) {
	itemIndex := uint32(id)
	item := &s.items[itemIndex]
	item.Position = dprec.Vec3{
		X: math.Inf(+1),
		Y: math.Inf(+1),
		Z: math.Inf(+1),
	}
	item.Radius = 1.0
	var zeroV T
	item.Value = zeroV
	s.dirtyItemIDs[itemIndex] = struct{}{}
	s.freeItemIDs.Push(itemIndex)
}

// VisitOverlapping finds all potentially intersecting item pairs in this set.
// It calls the specified visitor for each pair found.
func (s *SweepPruneSet[T]) VisitOverlapping(visitor PairVisitor[T]) {
	s.refresh()
	for candidate, count := range s.candidates {
		if count == 3 { // overlaps over X, Y, Z
			visitor.Visit(
				s.items[candidate.firstItemIndex].Value,
				s.items[candidate.secondItemIndex].Value,
			)
		}
	}
}

func (s *SweepPruneSet[T]) refresh() {
	if len(s.dirtyItemIDs) > 0 {
		s.updateMarkers()
		s.sortMarkers()
		s.determineCandidates()
	}
}

// updateMarkers ensures that markers have up-to-date positions.
func (s *SweepPruneSet[T]) updateMarkers() {
	for i := range s.seriesX {
		marker := &s.seriesX[i]
		itemIndex := marker.ItemIndex
		if _, ok := s.dirtyItemIDs[itemIndex]; ok {
			marker.Coord = s.items[itemIndex].BoundaryX(marker.IsStart)
		}
	}
	for i := range s.seriesY {
		marker := &s.seriesY[i]
		itemIndex := marker.ItemIndex
		if _, ok := s.dirtyItemIDs[itemIndex]; ok {
			marker.Coord = s.items[itemIndex].BoundaryY(marker.IsStart)
		}
	}
	for i := range s.seriesZ {
		marker := &s.seriesZ[i]
		itemIndex := marker.ItemIndex
		if _, ok := s.dirtyItemIDs[itemIndex]; ok {
			marker.Coord = s.items[itemIndex].BoundaryZ(marker.IsStart)
		}
	}
	maps.Clear(s.dirtyItemIDs)
}

func (s *SweepPruneSet[T]) sortMarkers() {
	sweepPruneMarkerList(s.seriesX).Sort()
	sweepPruneMarkerList(s.seriesY).Sort()
	sweepPruneMarkerList(s.seriesZ).Sort()
}

func (s *SweepPruneSet[T]) determineCandidates() {
	maps.Clear(s.candidates)
	s.addCandidatesForSeries(s.seriesX)
	s.addCandidatesForSeries(s.seriesY)
	s.addCandidatesForSeries(s.seriesZ)
}

func (s *SweepPruneSet[T]) addCandidatesForSeries(series []sweepPruneMarker) {
	maps.Clear(s.active)
	for _, marker := range s.seriesX {
		if math.IsInf(marker.Coord, 0) {
			break // we have reached deleted items
		}
		if marker.IsStart {
			for activeItemIndex := range s.active {
				pair := sortedSweepPrunePair(marker.ItemIndex, activeItemIndex)
				s.candidates[pair] = s.candidates[pair] + 1
			}
			s.active[marker.ItemIndex] = struct{}{}
		} else {
			delete(s.active, marker.ItemIndex)
		}
	}
}

func sortedSweepPrunePair(first, second uint32) sweepPrunePair {
	if first < second {
		return sweepPrunePair{
			firstItemIndex:  first,
			secondItemIndex: second,
		}
	} else {
		return sweepPrunePair{
			firstItemIndex:  second,
			secondItemIndex: first,
		}
	}
}

type sweepPrunePair struct {
	firstItemIndex  uint32
	secondItemIndex uint32
}

type sweepPruneMarker struct {
	Coord     float64
	ItemIndex uint32
	IsStart   bool
}

type sweepPruneMarkerList []sweepPruneMarker

func (l sweepPruneMarkerList) Sort() {
	slices.SortFunc(l, func(a, b sweepPruneMarker) int {
		return int(a.Coord - b.Coord)
	})
}

type sweepPruneItem[T any] struct {
	Position dprec.Vec3
	Radius   float64
	Value    T
}

func (i sweepPruneItem[T]) BoundaryX(isStart bool) float64 {
	if isStart {
		return i.Position.X - i.Radius
	} else {
		return i.Position.X + i.Radius
	}
}

func (i sweepPruneItem[T]) BoundaryY(isStart bool) float64 {
	if isStart {
		return i.Position.Y - i.Radius
	} else {
		return i.Position.Y + i.Radius
	}
}

func (i sweepPruneItem[T]) BoundaryZ(isStart bool) float64 {
	if isStart {
		return i.Position.Z - i.Radius
	} else {
		return i.Position.Z + i.Radius
	}
}

func (i sweepPruneItem[T]) IsDeleted() bool {
	return i.Position.IsInf()
}
