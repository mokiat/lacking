package animation

import (
	"cmp"
	"slices"
)

// NewBlend1DEntry creates a new Blend1DEntry.
func NewBlend1DEntry(coord float64, source Source) Blend1DEntry {
	return Blend1DEntry{
		Coord:  coord,
		Source: source,
	}
}

// Blend1DEntry specifies an animation to be used by a Blend1DSource.
type Blend1DEntry struct {
	Coord  float64
	Source Source
}

// NewBlend1DSource creates an animation source that can blend between
// multiple animations placed on a 1D line.
func NewBlend1DSource(entries ...Blend1DEntry) *Blend1DSource {
	if len(entries) == 0 {
		panic("at least one animation source is required")
	}
	slices.SortFunc(entries, func(a, b Blend1DEntry) int {
		return cmp.Compare(a.Coord, b.Coord)
	})
	return &Blend1DSource{
		entries:   entries,
		coord:     0.0,
		pairBlend: NewPairBlendSource(entries[0].Source, entries[0].Source),
		position:  0.0,
	}
}

var _ Source = (*Blend1DSource)(nil)

// Blend1DSource is an animation source that blends between the two closest
// animations positioned on a 1D line.
type Blend1DSource struct {
	entries   []Blend1DEntry
	coord     float64
	pairBlend *PairBlendSource
	position  float64
}

// Coord returns the blending coord.
func (s *Blend1DSource) Coord() float64 {
	return s.coord
}

// SetCoord changes the blending coord.
func (s *Blend1DSource) SetCoord(coord float64) {
	s.coord = coord

	lowerEntry := s.entries[0]
	for _, entry := range s.entries {
		if entry.Coord > coord {
			break
		}
		lowerEntry = entry
	}

	upperEntry := s.entries[len(s.entries)-1]
	for _, entry := range slices.Backward(s.entries) {
		if entry.Coord < coord {
			break
		}
		upperEntry = entry
	}

	factor := 0.0
	if lowerEntry != upperEntry {
		factor = (coord - lowerEntry.Coord) / (upperEntry.Coord - lowerEntry.Coord)
	}

	s.pairBlend.SetFirst(lowerEntry.Source)
	s.pairBlend.SetSecond(upperEntry.Source)
	s.pairBlend.SetFactor(factor)
	s.pairBlend.SetPosition(s.position)
}

// Synchronized returns whether animation synchronization should be used.
func (s *Blend1DSource) Synchronized() bool {
	return s.pairBlend.Synchronized()
}

// SetSynchronized sets whether animation synchronization should be used.
func (s *Blend1DSource) SetSynchronized(synchronized bool) {
	s.pairBlend.SetSynchronized(synchronized)
}

// Length returns the length of the animation in seconds.
func (s *Blend1DSource) Length() float64 {
	return s.pairBlend.Length()
}

// Position returns the current position of the animation in seconds.
func (s *Blend1DSource) Position() float64 {
	return s.pairBlend.Position()
}

// SetPosition sets the current position of the animation in seconds.
func (s *Blend1DSource) SetPosition(position float64) {
	s.position = position
	s.pairBlend.SetPosition(position)
}

// NodeTransform returns the transformation of the node with the specified
// name. The transformation is a blend of the transformations of the two
// sources of the node.
func (s *Blend1DSource) NodeTransform(name string) NodeTransform {
	return s.pairBlend.NodeTransform(name)
}
