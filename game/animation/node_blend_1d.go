package animation

import (
	"cmp"
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

var _ Node = (*Blend1DNode)(nil)

// Coord returns the blending coord.
func (n *Blend1DNode) Coord() float64 {
	return n.coord
}

// SetCoord changes the blending coord.
func (n *Blend1DNode) SetCoord(coord float64) {
	n.coord = coord

	lowerEntry := n.entries[0]
	for _, entry := range n.entries {
		if entry.Coord > coord {
			break
		}
		lowerEntry = entry
	}
	n.lower = lowerEntry.Node

	upperEntry := n.entries[len(n.entries)-1]
	for _, entry := range slices.Backward(n.entries) {
		if entry.Coord < coord {
			break
		}
		upperEntry = entry
	}
	n.upper = upperEntry.Node

	n.factor = 0.0
	if lowerEntry != upperEntry {
		n.factor = (coord - lowerEntry.Coord) / (upperEntry.Coord - lowerEntry.Coord)
	}
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *Blend1DNode) Reset() {
	n.SetFraction(n.Fraction())

	for _, entry := range n.entries {
		entry.Node.Reset()
	}
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *Blend1DNode) Rate() float64 {
	lowerRate := n.lower.Rate()
	upperRate := n.upper.Rate()
	// NOTE: The rates are flipped in the denominator on purpose. This is how
	// the math ends up if you derive this from lengths.
	return lowerRate * upperRate / dprec.Mix(upperRate, lowerRate, n.factor)
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *Blend1DNode) Fraction() float64 {
	return wrapFraction(n.progress)
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *Blend1DNode) SetFraction(fraction float64) {
	n.progress = fraction
	for _, entry := range n.entries {
		entry.Node.SetFraction(n.progress)
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *Blend1DNode) Advance(seconds, synchronizationRate float64) {
	rate := n.Rate()
	n.progress += rate * seconds * synchronizationRate
	n.progress = wrapFraction(n.progress)

	for _, entry := range n.entries {
		node := entry.Node
		adjustedRate := rate / node.Rate()
		node.Advance(seconds, synchronizationRate*adjustedRate)
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *Blend1DNode) BoneTransform(bone string) NodeTransform {
	lowerTransform := n.lower.BoneTransform(bone)
	upperTransform := n.upper.BoneTransform(bone)
	return BlendNodeTransforms(lowerTransform, upperTransform, n.factor)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *Blend1DNode) BoneTransformDelta(bone string) NodeTransform {
	lowerTransform := n.lower.BoneTransformDelta(bone)
	upperTransform := n.upper.BoneTransformDelta(bone)
	return BlendNodeTransforms(lowerTransform, upperTransform, n.factor)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *Blend1DNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	lowerTransform := n.lower.BoneTransformInterpolation(bone, fraction)
	upperTransform := n.upper.BoneTransformInterpolation(bone, fraction)
	return BlendNodeTransforms(lowerTransform, upperTransform, n.factor)
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
