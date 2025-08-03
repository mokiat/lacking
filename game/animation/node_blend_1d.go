package animation

import (
	"cmp"
	"math"
	"slices"

	"github.com/mokiat/gomath/dprec"
)

// NewBlend1DNode creates an animation node that can blend between
// multiple animations placed on a 1D line.
//
// NOTE: All animations are synchronized.
func NewBlend1DNode(entries ...Blend1DEntry) *Blend1DNode {
	if len(entries) == 0 {
		panic("at least one animation is required")
	}
	slices.SortFunc(entries, func(a, b Blend1DEntry) int {
		return cmp.Compare(a.Coord, b.Coord)
	})
	result := &Blend1DNode{
		entries:  entries,
		progress: 0.0,
	}
	result.SetCoord(0.0)
	return result
}

var _ Node = (*Blend1DNode)(nil)

// Blend1DNode is an animation source that blends between the two closest
// animations positioned on a 1D line.
//
// NOTE: All animations are considered to loop and are synchronized.
type Blend1DNode struct {
	entries  []Blend1DEntry
	progress float64
	coord    float64

	lower  Node
	upper  Node
	factor float64
}

// Coord returns the blending coord.
func (s *Blend1DNode) Coord() float64 {
	return s.coord
}

// SetCoord changes the blending coord.
func (s *Blend1DNode) SetCoord(coord float64) {
	s.coord = coord

	lowerEntry := s.entries[0]
	for _, entry := range s.entries {
		if entry.Coord > coord {
			break
		}
		lowerEntry = entry
	}
	s.lower = lowerEntry.Node

	upperEntry := s.entries[len(s.entries)-1]
	for _, entry := range slices.Backward(s.entries) {
		if entry.Coord < coord {
			break
		}
		upperEntry = entry
	}
	s.upper = upperEntry.Node

	s.factor = 0.0
	if lowerEntry != upperEntry {
		s.factor = (coord - lowerEntry.Coord) / (upperEntry.Coord - lowerEntry.Coord)
	}
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (s *Blend1DNode) Reset() {
	_, fraction := math.Modf(s.progress)
	s.Seek(fraction)

	for _, entry := range s.entries {
		entry.Node.Reset()
	}
}

// Rate returns the fraction of the animation length that advances each
// second.
func (s *Blend1DNode) Rate() float64 {
	lowerRate := s.lower.Rate()
	upperRate := s.upper.Rate()
	// NOTE: The rates are flipped in the denominator on purpose. This is how
	// the math ends up if you derive this from lengths.
	return lowerRate * upperRate / dprec.Mix(upperRate, lowerRate, s.factor)
}

// Seek relocates the animation to the specified position (fractional).
//
// NOTE: This resets the animation and accumulated delta is lost.
func (s *Blend1DNode) Seek(fraction float64) {
	s.progress = fraction
	for _, entry := range s.entries {
		entry.Node.Seek(s.progress)
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (s *Blend1DNode) Advance(seconds, synchronizationRate float64) {
	rate := s.Rate()
	s.progress += rate * seconds * synchronizationRate

	for _, entry := range s.entries {
		node := entry.Node
		adjustedRate := rate / node.Rate()
		node.Advance(seconds, synchronizationRate*adjustedRate)
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (s *Blend1DNode) BoneTransform(bone string) NodeTransform {
	lowerTransform := s.lower.BoneTransform(bone)
	upperTransform := s.upper.BoneTransform(bone)
	return BlendNodeTransforms(lowerTransform, upperTransform, s.factor)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (s *Blend1DNode) BoneTransformDelta(bone string) NodeTransform {
	lowerTransform := s.lower.BoneTransformDelta(bone)
	upperTransform := s.upper.BoneTransformDelta(bone)
	return BlendNodeTransforms(lowerTransform, upperTransform, s.factor)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (s *Blend1DNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	lowerTransform := s.lower.BoneTransformInterpolation(bone, fraction)
	upperTransform := s.upper.BoneTransformInterpolation(bone, fraction)
	return BlendNodeTransforms(lowerTransform, upperTransform, s.factor)
}

// NewBlend1DEntry creates a new Blend1DEntry.
func NewBlend1DEntry(coord float64, node Node) Blend1DEntry {
	return Blend1DEntry{
		Coord: coord,
		Node:  node,
	}
}

// Blend1DEntry specifies an animation to be used by a Blend1DSource.
type Blend1DEntry struct {
	Coord float64
	Node  Node
}
