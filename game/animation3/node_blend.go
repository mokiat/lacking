package animation3

// BlendNode is a [Node] that blends two animations according to a weight. A
// weight of 0.0 shows the first animation, a weight of 1.0 shows the second
// animation and a weight in between shows a blend of the two. Both nodes are
// expected to be of the same kind (see [Node.Transform]), which is also the
// kind of the result.
//
// Unlike nodes that follow one of their children, a BlendNode keeps a
// playback position of its own. It advances at a rate that is blended from
// the rates of its synchronized children, so that synchronized children of
// different lengths, such as a walk and a run cycle, stay in phase while the
// length of the blended cycle changes smoothly with the weight. Each child
// is advanced in real time and, if it is synchronized, it is then kept at the
// fraction of the BlendNode. Children that are not synchronized play at
// their own pace and do not affect the rate.
//
// The fraction of a BlendNode wraps around when it reaches either end, so
// the node is meant for cyclic animations. A synchronized child that does not
// loop is moved back to its start when the fraction wraps around.
type BlendNode struct {
	synchronizationAspect
	first    Node
	second   Node
	weight   float64
	fraction float64
}

var _ Node = (*BlendNode)(nil)

// NewBlendNode creates a new [BlendNode] that blends the specified first and
// second nodes, with a weight of 0.0 and a fraction of 0.0. The BlendNode
// takes control of the playback of both nodes, which should not be advanced
// or relocated through other means afterwards. The node is not synchronized.
func NewBlendNode(first, second Node) *BlendNode {
	return &BlendNode{
		first:  first,
		second: second,
	}
}

// Blend is a shorthand for [NewBlendNode], which is useful during node tree
// construction.
func Blend(first, second Node) *BlendNode {
	return NewBlendNode(first, second)
}

// Weight returns the blend weight between the first and the second node.
func (n *BlendNode) Weight() float64 {
	return n.weight
}

// SetWeight changes the blend weight between the first and the second node.
// A weight of 0.0 shows the first node and a weight of 1.0 shows the second
// node. The weight is clamped into the range [0.0..1.0].
func (n *BlendNode) SetWeight(weight float64) {
	n.weight = clampFraction(weight)
}

// Rate returns the rate at which the node advances, which is blended from
// the rates of the synchronized children that have a rate greater than 0.0.
// The rates are blended so that the length of the cycle, rather than the
// rate itself, changes linearly with the weight. If only one child takes
// part, its rate is returned unchanged. If no child takes part, the node has
// no meaningful length and the rate is 0.0.
func (n *BlendNode) Rate() float64 {
	return blendRates(n.first, n.second, n.weight)
}

// Fraction returns the playback position of the node, in the range
// [0.0..1.0).
func (n *BlendNode) Fraction() float64 {
	return n.fraction
}

// SetFraction relocates the node to the specified fraction, which is wrapped
// into the range [0.0..1.0). Synchronized children are relocated to the same
// fraction, whereas the others keep their playback position.
func (n *BlendNode) SetFraction(fraction float64) {
	n.fraction = wrapFraction(fraction)
	if n.first.IsSynchronized() {
		n.first.SetFraction(n.fraction)
	}
	if n.second.IsSynchronized() {
		n.second.SetFraction(n.fraction)
	}
}

// Advance moves the node by the specified delta seconds at its [BlendNode.Rate],
// wrapping around at either end. If the rate is 0.0, the fraction of the node
// does not change. Both children are advanced by the same delta seconds and
// synchronized children are then snapped to the fraction of the node.
func (n *BlendNode) Advance(deltaSeconds float64) {
	n.fraction = wrapFraction(n.fraction + n.Rate()*deltaSeconds)
	n.first.Advance(deltaSeconds)
	if n.first.IsSynchronized() {
		n.first.SetFraction(n.fraction)
	}
	n.second.Advance(deltaSeconds)
	if n.second.IsSynchronized() {
		n.second.SetFraction(n.fraction)
	}
}

// Transform returns the transformation of the specified bone, blended
// through [BlendTransforms] from that of the first node towards that of the
// second node by the weight.
//
// The bone must be a valid index in the skeleton of the child nodes.
func (n *BlendNode) Transform(bone int) Transform {
	switch n.weight {
	case 0.0:
		return n.first.Transform(bone)
	case 1.0:
		return n.second.Transform(bone)
	default:
		return BlendTransforms(n.first.Transform(bone), n.second.Transform(bone), n.weight)
	}
}

// DeltaTransform returns the motion that the specified bone experiences
// throughout the next delta seconds, blended through [BlendTransforms] from
// that of the first node towards that of the second node by the weight.
//
// For a weight between 0.0 and 1.0 the result approximates the motion of the
// blended transformation. The motion of a synchronized child is measured in
// real time, so it is approximate when the rate of the child differs from
// that of the node.
//
// The bone must be a valid index in the skeleton of the child nodes.
func (n *BlendNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	switch n.weight {
	case 0.0:
		return n.first.DeltaTransform(bone, deltaSeconds)
	case 1.0:
		return n.second.DeltaTransform(bone, deltaSeconds)
	default:
		return BlendTransforms(n.first.DeltaTransform(bone, deltaSeconds), n.second.DeltaTransform(bone, deltaSeconds), n.weight)
	}
}
