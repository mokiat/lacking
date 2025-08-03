package animation

import "github.com/mokiat/gog/filter"

// NewMaskNode creates a new animation node that picks specific bones
// from the specified node.
func NewMaskNode(delegate Node, selection filter.Func[string]) *MaskNode {
	return &MaskNode{
		delegate:  delegate,
		selection: selection,
	}
}

var _ Node = (*MaskNode)(nil)

// MaskNode is an animation source that picks specific bones
// from another animation source.
type MaskNode struct {
	delegate  Node
	selection filter.Func[string]
}

// Rate returns the fraction of the animation length that advances each
// second.
func (s *MaskNode) Rate() float64 {
	return s.delegate.Rate()
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (s *MaskNode) Reset() {
	s.delegate.Reset()
}

// Progress returns the current fraction of the animation that has
// advanced since the start.
//
// This value will always be in the range [0.0..1.0).
func (s *MaskNode) Progress() float64 {
	return s.delegate.Progress()
}

// SetProgress changes the current position of the animation to the
// specified fraction.
//
// It is possible to set this value above 1.0, and in fact is necessary
// during update, so that it can handle loops and interpolation correctly,
// as setting the value directly to the wrapped-around value might indicate
// a reverse animation or a fractional animation.
//
// Internally, once applied, the progress will be normalized to [0.0..1.0).
func (s *MaskNode) SetProgress(fraction float64) {
	s.delegate.SetProgress(fraction)
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (s *MaskNode) BoneTransform(bone string) NodeTransform {
	if !s.selection(bone) {
		return NodeTransform{}
	}
	return s.delegate.BoneTransform(bone)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (s *MaskNode) BoneTransformDelta(bone string) NodeTransform {
	if !s.selection(bone) {
		return NodeTransform{}
	}
	return s.delegate.BoneTransformDelta(bone)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (s *MaskNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	if !s.selection(bone) {
		return NodeTransform{}
	}
	return s.delegate.BoneTransformInterpolation(bone, fraction)
}
