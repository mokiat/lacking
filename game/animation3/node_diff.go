package animation3

// DiffNode is a [Node] that produces the difference of a source animation
// relative to a reference animation. Both nodes are expected to be absolute
// (see [Node.Transform]) and the result is additive, meant to be applied on
// top of another animation through an [AddNode]. Applying the result on top
// of the reference reproduces the source.
//
// The reference is commonly a static pose, such as the first frame of the
// source animation produced by a [ClipPoseNode], so that the result holds
// only the motion of the source animation.
//
// The source node drives the playback of the DiffNode: its rate and fraction
// are those of the source node. The reference node is advanced in real time
// alongside the source node and, if it is synchronized, it is kept at the
// same fraction as the source node.
//
// The synchronization setting of the source node has no effect, since the
// source node defines the fraction of the DiffNode. Whether the source node
// is kept in phase with the parent of the DiffNode is decided by the
// synchronization setting of the DiffNode itself.
type DiffNode struct {
	synchronizationAspect
	source    Node
	reference Node
}

var _ Node = (*DiffNode)(nil)

// NewDiffNode creates a new [DiffNode] that produces the difference of the
// specified source node relative to the specified reference node. The
// DiffNode takes control of the playback of both nodes, which should not be
// advanced or relocated through other means afterwards. The node is not
// synchronized.
func NewDiffNode(source, reference Node) *DiffNode {
	return &DiffNode{
		source:    source,
		reference: reference,
	}
}

// Diff is a shorthand for [NewDiffNode], which is useful during node tree
// construction.
func Diff(source, reference Node) *DiffNode {
	return NewDiffNode(source, reference)
}

// Rate returns the rate of the source node.
func (n *DiffNode) Rate() float64 {
	return n.source.Rate()
}

// Fraction returns the fraction of the source node.
func (n *DiffNode) Fraction() float64 {
	return n.source.Fraction()
}

// SetFraction relocates the source node to the specified fraction. A
// synchronized reference node is relocated to the resulting fraction of the
// source node, whereas a non-synchronized one keeps its playback position.
func (n *DiffNode) SetFraction(fraction float64) {
	n.source.SetFraction(fraction)
	if n.reference.IsSynchronized() {
		n.reference.SetFraction(n.source.Fraction())
	}
}

// Advance moves both the source and the reference node by the specified
// delta seconds. A synchronized reference node is then snapped to the
// fraction of the source node.
func (n *DiffNode) Advance(seconds float64) {
	n.source.Advance(seconds)
	n.reference.Advance(seconds)
	if n.reference.IsSynchronized() {
		n.reference.SetFraction(n.source.Fraction())
	}
}

// Transform returns the difference, as produced by [DiffTransforms], of the
// transformation of the specified bone in the source node relative to that
// in the reference node. A bone that is in the same transformation in both
// nodes, such as one that neither of them animates, returns the
// [IdentityTransform].
//
// The bone must be a valid index in the skeleton of the child nodes.
func (n *DiffNode) Transform(bone int) Transform {
	return DiffTransforms(n.source.Transform(bone), n.reference.Transform(bone))
}

// DeltaTransform returns the [IdentityTransform], since an additive
// difference does not contribute to root motion.
func (n *DiffNode) DeltaTransform(bone int, delta float64) Transform {
	return IdentityTransform()
}
