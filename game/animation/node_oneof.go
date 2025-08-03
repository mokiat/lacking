package animation

// NewOneOfNode creates an animation node that produces one specific animation
// from a set of animations.
func NewOneOfNode(animations map[string]Node) *OneOfNode {
	return &OneOfNode{
		animations:          animations,
		activeAnimationName: "",
		activeAnimation:     nil,
	}
}

var _ Node = (*OneOfNode)(nil)

// OneOfNode is an animation node that plays one specific animation
// from a set of animations. The active animation can be changed at any time.
type OneOfNode struct {
	animations          map[string]Node
	activeAnimationName string
	activeAnimation     Node
}

// ActiveAnimation returns the underlying node that is currently being used.
func (n *OneOfNode) ActiveAnimation() Node {
	return n.activeAnimation
}

// PickAnimation changes to the specified animation.
//
// The reset flag controls whether the new animation should start from the
// beginning.
func (n *OneOfNode) PickAnimation(name string, reset bool) {
	n.activeAnimationName = name
	n.activeAnimation = n.animations[name]
	if reset && (n.activeAnimation != nil) {
		n.activeAnimation.Seek(0.0)
		n.activeAnimation.Reset()
	}
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *OneOfNode) Reset() {
	for _, node := range n.animations {
		node.Reset()
	}
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *OneOfNode) Rate() float64 {
	if n.activeAnimation == nil {
		return 1.0
	}
	return n.activeAnimation.Rate()
}

// Seek relocates the animation to the specified position (fractional).
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *OneOfNode) Seek(fraction float64) {
	if n.activeAnimation != nil {
		n.activeAnimation.Seek(fraction)
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *OneOfNode) Advance(seconds, synchronizationRate float64) {
	if n.activeAnimation != nil {
		n.activeAnimation.Advance(seconds, synchronizationRate)
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *OneOfNode) BoneTransform(bone string) NodeTransform {
	if n.activeAnimation == nil {
		return NodeTransform{}
	}
	return n.activeAnimation.BoneTransform(bone)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *OneOfNode) BoneTransformDelta(bone string) NodeTransform {
	if n.activeAnimation == nil {
		return NodeTransform{}
	}
	return n.activeAnimation.BoneTransformDelta(bone)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *OneOfNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	if n.activeAnimation == nil {
		return NodeTransform{}
	}
	return n.activeAnimation.BoneTransformInterpolation(bone, fraction)
}
