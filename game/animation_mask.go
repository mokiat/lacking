package game

import "github.com/mokiat/gog/filter"

// NewAnimationMask creates a new AnimationSource that picks specific bones
// from the specified AnimationSource.
func NewAnimationMask(delegate AnimationSource, selection filter.Func[string]) *AnimationMask {
	return &AnimationMask{
		AnimationSource: delegate,
		selection:       selection,
	}
}

var _ AnimationSource = (*AnimationMask)(nil)

// AnimationMask is an animation source that picks specific bones
// from another animation source.
type AnimationMask struct {
	AnimationSource
	selection filter.Func[string]
}

// NodeTransform returns the transformation of the node with the specified
// name at the current time position.
func (m *AnimationMask) NodeTransform(name string) NodeTransform {
	if !m.selection(name) {
		return NodeTransform{}
	}
	return m.AnimationSource.NodeTransform(name)
}
