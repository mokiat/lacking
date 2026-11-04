package animation

// MaskedBlendNode is a [Node] that blends a layer animation over a base
// animation, bone by bone, according to a mask. The mask assigns a weight to
// each bone: a bone with a weight of 0.0 shows the base animation, a bone
// with a weight of 1.0 shows the layer animation, and a bone in between
// shows a blend of the two. This makes it possible, for example, to play a
// different animation on the upper body while the legs keep playing the
// base animation. Both nodes are expected to be of the same kind (see
// [Node.Transform]), which is also the kind of the result.
//
// An overall weight scales the weights of all bones in the mask, which
// allows the layer to be faded in and out as a whole.
//
// The base node drives the playback of the MaskedBlendNode: its rate and
// fraction are those of the base node. The layer node is advanced in real
// time alongside the base node and, if it is synchronized, it is kept at the
// same fraction as the base node.
//
// The synchronization setting of the base node has no effect, since the base
// node defines the fraction of the MaskedBlendNode. Whether the base node is
// kept in phase with the parent of the MaskedBlendNode is decided by the
// synchronization setting of the MaskedBlendNode itself.
type MaskedBlendNode struct {
	synchronizationAspect
	base        Node
	layer       Node
	boneWeights []float64
	weight      float64
}

var _ Node = (*MaskedBlendNode)(nil)

// NewMaskedBlendNode creates a new [MaskedBlendNode] that blends the
// specified layer node over the specified base node, according to the
// specified bone weights keyed by bone name, with an overall weight of 1.0.
// Bone weights are resolved against the specified skeleton: bones without an
// entry have a weight of 0.0 and show the base animation, and entries for
// bones that are not in the skeleton are ignored. The node does not retain
// the specified map.
//
// The MaskedBlendNode takes control of the playback of both nodes, which
// should not be advanced or relocated through other means afterwards. The
// node is not synchronized.
func NewMaskedBlendNode(skeleton *Skeleton, base, layer Node, boneWeights map[string]float64) *MaskedBlendNode {
	weights := make([]float64, skeleton.BoneCount())
	for i := range weights {
		weights[i] = boneWeights[skeleton.BoneName(i)]
	}
	return &MaskedBlendNode{
		base:        base,
		layer:       layer,
		boneWeights: weights,
		weight:      1.0,
	}
}

// Weight returns the overall weight that scales the weights of all bones in
// the mask.
func (n *MaskedBlendNode) Weight() float64 {
	return n.weight
}

// SetWeight changes the overall weight that scales the weights of all bones
// in the mask. A weight of 1.0 applies the mask as configured and a weight of
// 0.0 shows the base animation on all bones. The weight is not clamped and
// values that take the effective weight of a bone outside the range
// [0.0..1.0] extrapolate, as by [BlendTransforms].
func (n *MaskedBlendNode) SetWeight(weight float64) {
	n.weight = weight
}

// Rate returns the rate of the base node.
func (n *MaskedBlendNode) Rate() float64 {
	return n.base.Rate()
}

// Fraction returns the fraction of the base node.
func (n *MaskedBlendNode) Fraction() float64 {
	return n.base.Fraction()
}

// SetFraction relocates the base node to the specified fraction. A
// synchronized layer node is relocated to the resulting fraction of the base
// node, whereas a non-synchronized one keeps its playback position.
func (n *MaskedBlendNode) SetFraction(fraction float64) {
	n.base.SetFraction(fraction)
	if n.layer.IsSynchronized() {
		n.layer.SetFraction(n.base.Fraction())
	}
}

// Advance moves both the base and the layer node by the specified delta
// seconds. A synchronized layer node is then snapped to the fraction of the
// base node.
func (n *MaskedBlendNode) Advance(deltaSeconds float64) {
	n.base.Advance(deltaSeconds)
	n.layer.Advance(deltaSeconds)
	if n.layer.IsSynchronized() {
		n.layer.SetFraction(n.base.Fraction())
	}
}

// Transform returns the transformation of the specified bone, blended
// through [BlendTransforms] from that of the base node towards that of the
// layer node by the effective weight of the bone, which is its mask weight
// scaled by the overall weight.
//
// The bone must be a valid index in the skeleton of the node.
func (n *MaskedBlendNode) Transform(bone int) Transform {
	switch weight := n.boneWeight(bone); weight {
	case 0.0:
		return n.base.Transform(bone)
	case 1.0:
		return n.layer.Transform(bone)
	default:
		return BlendTransforms(n.base.Transform(bone), n.layer.Transform(bone), weight)
	}
}

// DeltaTransform returns the motion that the specified bone experiences
// throughout the next delta seconds, blended through [BlendTransforms] from
// that of the base node towards that of the layer node by the effective
// weight of the bone. A bone with an effective weight of 0.0 moves as in the
// base node, so a mask that excludes the root bone leaves root motion
// entirely to the base node.
//
// For an effective weight between 0.0 and 1.0 the result approximates the
// motion of the blended transformation. The motion of a synchronized layer
// node is measured in real time, so it is approximate when the layer node is
// kept in phase with a base node of a different rate.
//
// The bone must be a valid index in the skeleton of the node.
func (n *MaskedBlendNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	switch weight := n.boneWeight(bone); weight {
	case 0.0:
		return n.base.DeltaTransform(bone, deltaSeconds)
	case 1.0:
		return n.layer.DeltaTransform(bone, deltaSeconds)
	default:
		return BlendTransforms(n.base.DeltaTransform(bone, deltaSeconds), n.layer.DeltaTransform(bone, deltaSeconds), weight)
	}
}

// boneWeight returns the effective weight of the specified bone.
func (n *MaskedBlendNode) boneWeight(bone int) float64 {
	return n.boneWeights[bone] * n.weight
}
