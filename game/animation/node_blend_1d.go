package animation

import (
	"cmp"
	"slices"
)

// NewBlend1DNode creates an animation node that can blend between
// multiple animations placed on a 1D line.
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
	result.SetBlendCoord(0.0)
	return result
}

// Blend1DNode is an animation source that blends between the two closest
// animations positioned on a 1D line.
//
// NOTE: All animations are considered to loop and are synchronized.
type Blend1DNode struct {
	entries    []Blend1DEntry
	progress   float64
	blendCoord float64

	lower  Node
	upper  Node
	factor float64

	synchronized bool
}

var _ Node = (*Blend1DNode)(nil)

// BlendCoord returns the blending coord.
func (n *Blend1DNode) BlendCoord() float64 {
	return n.blendCoord
}

// SetBlendCoord changes the blending coord.
func (n *Blend1DNode) SetBlendCoord(blendCoord float64) {
	n.blendCoord = blendCoord
	n.lower, n.upper, n.factor = n.resolveCoord(blendCoord)
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *Blend1DNode) Rate() float64 {
	return blendRates(n.lower, n.upper, n.factor)
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
	n.progress = wrapFraction(fraction)
	for _, entry := range n.entries {
		node := entry.Node
		if node.IsSynchronized() {
			node.SetFraction(n.progress)
		}
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
		if node.IsSynchronized() {
			adjustedRate := rate / node.Rate()
			node.Advance(seconds, synchronizationRate*adjustedRate)
		} else {
			node.Advance(seconds, 1.0)
		}
	}
}

// IsSynchronized returns whether the node should be synchronized.
func (n *Blend1DNode) IsSynchronized() bool {
	return n.synchronized
}

// SetSynchronized configures whether the node should be synchronized.
func (n *Blend1DNode) SetSynchronized(synchronized bool) {
	n.synchronized = synchronized
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *Blend1DNode) Synchronize() {
	for _, entry := range n.entries {
		node := entry.Node
		if node.IsSynchronized() {
			node.SetFraction(n.progress)
		}
		node.Synchronize()
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

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *Blend1DNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	lowerTransform := n.lower.BoneDeltaTransform(bone, delta)
	upperTransform := n.upper.BoneDeltaTransform(bone, delta)
	return BlendNodeTransforms(lowerTransform, upperTransform, n.factor)
}

func (n *Blend1DNode) resolveCoord(coord float64) (Node, Node, float64) {
	lowerEntry := n.entries[0]
	for _, entry := range n.entries {
		if entry.Coord > coord {
			break
		}
		lowerEntry = entry
	}

	upperEntry := n.entries[len(n.entries)-1]
	for _, entry := range slices.Backward(n.entries) {
		if entry.Coord < coord {
			break
		}
		upperEntry = entry
	}

	factor := 0.0
	if lowerEntry != upperEntry {
		factor = (coord - lowerEntry.Coord) / (upperEntry.Coord - lowerEntry.Coord)
	}

	return lowerEntry.Node, upperEntry.Node, factor
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
