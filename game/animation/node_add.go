package animation

// AddNode is a [Node] that applies an additive layer on top of a base
// animation. The base node is expected to be absolute and the additive node
// is expected to be additive (see [Node.Transform]), for example the output
// of a node that computes the difference between two animations. The result
// is absolute.
//
// The strength of the additive layer is controlled through a weight. A
// weight of 1.0 applies the layer in full and a weight of 0.0 leaves the
// base animation unchanged.
//
// The base node drives the playback of the AddNode: its rate and fraction
// are those of the base node. The additive node is advanced in real time
// alongside the base node and, if it is synchronized, it is kept at the same
// fraction as the base node.
//
// The synchronization setting of the base node has no effect, since the base
// node defines the fraction of the AddNode. Whether the base node is kept in
// phase with the parent of the AddNode is decided by the synchronization
// setting of the AddNode itself.
type AddNode struct {
	synchronizationAspect
	base     Node
	additive Node
	weight   float64
}

var _ Node = (*AddNode)(nil)

// NewAddNode creates a new [AddNode] that applies the specified additive node
// on top of the specified base node with a weight of 1.0. The AddNode takes
// control of the playback of both nodes, which should not be advanced or
// relocated through other means afterwards. The node is not synchronized.
func NewAddNode(base, additive Node) *AddNode {
	return &AddNode{
		base:     base,
		additive: additive,
		weight:   1.0,
	}
}

// Add is a shorthand for [NewAddNode], which is useful during node tree
// construction.
func Add(base, additive Node) *AddNode {
	return NewAddNode(base, additive)
}

// Weight returns the strength with which the additive layer is applied.
func (n *AddNode) Weight() float64 {
	return n.weight
}

// SetWeight changes the strength with which the additive layer is applied.
// A weight of 1.0 applies the layer in full and a weight of 0.0 leaves the
// base animation unchanged. Values in between apply a proportional amount of
// the layer, as produced by [BlendTransforms] between the
// [IdentityTransform] and the additive transformation. The weight is not
// clamped and values outside the range [0.0..1.0] extrapolate.
func (n *AddNode) SetWeight(weight float64) {
	n.weight = weight
}

// Rate returns the rate of the base node.
func (n *AddNode) Rate() float64 {
	return n.base.Rate()
}

// Fraction returns the fraction of the base node.
func (n *AddNode) Fraction() float64 {
	return n.base.Fraction()
}

// SetFraction relocates the base node to the specified fraction. A
// synchronized additive node is relocated to the resulting fraction of the
// base node, whereas a non-synchronized one keeps its playback position.
func (n *AddNode) SetFraction(fraction float64) {
	n.base.SetFraction(fraction)
	if n.additive.IsSynchronized() {
		n.additive.SetFraction(n.base.Fraction())
	}
}

// Advance moves both the base and the additive node by the specified delta
// seconds. A synchronized additive node is then snapped to the fraction of
// the base node.
func (n *AddNode) Advance(deltaSeconds float64) {
	n.base.Advance(deltaSeconds)
	n.additive.Advance(deltaSeconds)
	if n.additive.IsSynchronized() {
		n.additive.SetFraction(n.base.Fraction())
	}
}

// Transform returns the transformation of the specified bone produced by the
// base node, with that of the additive node applied on top of it through
// [AddTransforms] according to the weight.
//
// The bone must be a valid index in the skeleton of the child nodes.
func (n *AddNode) Transform(bone int) Transform {
	baseTransform := n.base.Transform(bone)
	switch n.weight {
	case 0.0:
		return baseTransform
	case 1.0:
		return AddTransforms(baseTransform, n.additive.Transform(bone))
	default:
		additiveTransform := BlendTransforms(IdentityTransform(), n.additive.Transform(bone), n.weight)
		return AddTransforms(baseTransform, additiveTransform)
	}
}

// DeltaTransform returns the motion that the specified bone experiences in
// the base node throughout the next delta seconds. The additive layer does
// not contribute to root motion.
//
// The bone must be a valid index in the skeleton of the child nodes.
func (n *AddNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	return n.base.DeltaTransform(bone, deltaSeconds)
}
