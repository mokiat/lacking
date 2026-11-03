package animation3

import (
	"cmp"
	"slices"
)

// Blend1DEntry specifies a node that takes part in a [Blend1DNode] and the
// coordinate at which it is placed on the line of the blend.
type Blend1DEntry struct {

	// Coord is the coordinate of the node on the line of the blend.
	Coord float64

	// Node is the node that is shown at the coordinate.
	Node Node
}

// Blend1DNode is a [Node] that blends between a set of animations placed at
// coordinates along a line, such as walk, jog and run animations placed at
// the movement speeds that they correspond to. Setting the blend coordinate
// to a gameplay value, such as the current speed, shows a blend of the two
// entries that surround that coordinate. A coordinate before the first entry
// or after the last entry shows that entry alone. All nodes are expected to
// be of the same kind (see [Node.Transform]), which is also the kind of the
// result.
//
// Like a [BlendNode], a Blend1DNode keeps a playback position of its own and
// advances at a rate that is blended from the rates of the two active
// entries, as long as they are synchronized. All entries are advanced in real
// time, whether they are active or not, and synchronized entries are then
// kept at the fraction of the Blend1DNode. Entries that are not synchronized
// play at their own pace and do not affect the rate.
//
// The fraction of a Blend1DNode wraps around when it reaches either end, so
// the node is meant for cyclic animations. A synchronized entry that does
// not loop is moved back to its start when the fraction wraps around.
type Blend1DNode struct {
	synchronizationAspect
	entries  []Blend1DEntry
	coord    float64
	fraction float64

	lower  Node
	upper  Node
	factor float64
}

var _ Node = (*Blend1DNode)(nil)

// NewBlend1DNode creates a new [Blend1DNode] from the specified entries,
// which need not be sorted by coordinate. Entries that share a coordinate
// produce an abrupt switch between them at that coordinate. The node does not
// retain the specified slice.
//
// The blend coordinate starts at the coordinate of the first entry along the
// line and the fraction starts at 0.0. The Blend1DNode takes control of the
// playback of all nodes, which should not be advanced or relocated through
// other means afterwards. The node is not synchronized.
//
// At least one entry must be specified, otherwise the function panics.
func NewBlend1DNode(entries ...Blend1DEntry) *Blend1DNode {
	if len(entries) == 0 {
		panic("at least one entry is required")
	}
	entries = slices.Clone(entries)
	slices.SortStableFunc(entries, func(a, b Blend1DEntry) int {
		return cmp.Compare(a.Coord, b.Coord)
	})
	result := &Blend1DNode{
		entries: entries,
	}
	result.SetBlendCoord(entries[0].Coord)
	return result
}

// BlendCoord returns the coordinate along the line that determines which
// entries are shown and how they are blended.
func (n *Blend1DNode) BlendCoord() float64 {
	return n.coord
}

// SetBlendCoord changes the coordinate along the line that determines which
// entries are shown and how they are blended. A coordinate between two
// entries blends them linearly according to its distance to each of them,
// whereas a coordinate outside the range of the entries shows the closest
// entry alone.
func (n *Blend1DNode) SetBlendCoord(coord float64) {
	n.coord = coord

	index := 0
	for index < len(n.entries)-1 && n.entries[index+1].Coord <= coord {
		index++
	}
	lowerEntry := n.entries[index]
	if index == len(n.entries)-1 || coord <= lowerEntry.Coord {
		n.lower, n.upper, n.factor = lowerEntry.Node, lowerEntry.Node, 0.0
	} else {
		upperEntry := n.entries[index+1]
		factor := (coord - lowerEntry.Coord) / (upperEntry.Coord - lowerEntry.Coord)
		n.lower, n.upper, n.factor = lowerEntry.Node, upperEntry.Node, factor
	}
}

// Rate returns the rate at which the node advances, which is blended from
// the rates of the two active entries in the same way as by
// [BlendNode.Rate]. If no active entry is synchronized with a rate greater
// than 0.0, the node has no meaningful length and the rate is 0.0.
func (n *Blend1DNode) Rate() float64 {
	return blendRates(n.lower, n.upper, n.factor)
}

// Fraction returns the playback position of the node, in the range
// [0.0..1.0).
func (n *Blend1DNode) Fraction() float64 {
	return n.fraction
}

// SetFraction relocates the node to the specified fraction, which is wrapped
// into the range [0.0..1.0). All synchronized entries, active or not, are
// relocated to the same fraction, whereas the others keep their playback
// position.
func (n *Blend1DNode) SetFraction(fraction float64) {
	n.fraction = wrapFraction(fraction)
	for _, entry := range n.entries {
		if entry.Node.IsSynchronized() {
			entry.Node.SetFraction(n.fraction)
		}
	}
}

// Advance moves the node by the specified delta seconds at its
// [Blend1DNode.Rate], wrapping around at either end. If the rate is 0.0, the
// fraction of the node does not change. All entries, active or not, are
// advanced by the same delta seconds and synchronized ones are then snapped
// to the fraction of the node.
func (n *Blend1DNode) Advance(deltaSeconds float64) {
	n.fraction = wrapFraction(n.fraction + n.Rate()*deltaSeconds)
	for _, entry := range n.entries {
		entry.Node.Advance(deltaSeconds)
		if entry.Node.IsSynchronized() {
			entry.Node.SetFraction(n.fraction)
		}
	}
}

// Transform returns the transformation of the specified bone, blended
// through [BlendTransforms] between the two active entries according to the
// blend coordinate.
//
// The bone must be a valid index in the skeleton of the entry nodes.
func (n *Blend1DNode) Transform(bone int) Transform {
	switch {
	case n.factor == 0.0 || n.lower == n.upper:
		return n.lower.Transform(bone)
	case n.factor == 1.0:
		return n.upper.Transform(bone)
	default:
		return BlendTransforms(n.lower.Transform(bone), n.upper.Transform(bone), n.factor)
	}
}

// DeltaTransform returns the motion that the specified bone experiences
// throughout the next delta seconds, blended through [BlendTransforms]
// between the two active entries according to the blend coordinate.
//
// Between two entries the result approximates the motion of the blended
// transformation. The motion of a synchronized entry is measured in real
// time, so it is approximate when the rate of the entry differs from that of
// the node.
//
// The bone must be a valid index in the skeleton of the entry nodes.
func (n *Blend1DNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	switch {
	case n.factor == 0.0 || n.lower == n.upper:
		return n.lower.DeltaTransform(bone, deltaSeconds)
	case n.factor == 1.0:
		return n.upper.DeltaTransform(bone, deltaSeconds)
	default:
		return BlendTransforms(n.lower.DeltaTransform(bone, deltaSeconds), n.upper.DeltaTransform(bone, deltaSeconds), n.factor)
	}
}
