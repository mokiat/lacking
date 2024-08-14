package game

import "github.com/mokiat/gomath/dprec"

// NewPairAnimationBlending creates a new pair animation blending node
// with the specified sources.
func NewPairAnimationBlending(first, second AnimationSource) *PairAnimationBlending {
	return &PairAnimationBlending{
		first:        first,
		second:       second,
		synchronized: true,
	}
}

var _ AnimationSource = (*PairAnimationBlending)(nil)

// PairAnimationBlending represents an animation source that blends two
// animation sources. The blending factor is determined by the factor
// field of the node.
type PairAnimationBlending struct {
	first        AnimationSource
	second       AnimationSource
	synchronized bool
	factor       float64
}

// First returns the first source of the node.
func (n *PairAnimationBlending) First() AnimationSource {
	return n.first
}

// SetFirst sets the first source of the node.
func (n *PairAnimationBlending) SetFirst(first AnimationSource) {
	n.first = first
}

// Second returns the second source of the node.
func (n *PairAnimationBlending) Second() AnimationSource {
	return n.second
}

// SetSecond sets the second source of the node.
func (n *PairAnimationBlending) SetSecond(second AnimationSource) {
	n.second = second
}

// Synchronized returns whether the two sources of the node are synchronized.
func (n *PairAnimationBlending) Synchronized() bool {
	return n.synchronized
}

// SetSynchronized sets whether the two sources of the node are synchronized.
func (n *PairAnimationBlending) SetSynchronized(synchronized bool) {
	n.synchronized = synchronized
}

// Factor returns the blending factor of the node. A value of 0.0 means
// that the first source is used, a value of 1.0 means that the second
// source is used. The value is clamped to the range [0.0, 1.0].
func (n *PairAnimationBlending) Factor() float64 {
	return n.factor
}

// SetFactor sets the blending factor of the node. The value is clamped
// to the range [0.0, 1.0].
func (n *PairAnimationBlending) SetFactor(factor float64) {
	n.factor = dprec.Clamp(factor, 0.0, 1.0)
}

// Length returns the length of the animation in seconds.
func (n *PairAnimationBlending) Length() float64 {
	lngFirst := n.first.Length()
	lngSecond := n.second.Length()
	if n.synchronized {
		return dprec.Mix(lngFirst, lngSecond, n.factor)
	} else {
		return max(lngFirst, lngSecond)
	}
}

// Position returns the current position of the animation in seconds.
func (n *PairAnimationBlending) Position() float64 {
	lngFirst := n.first.Length()
	lngSecond := n.second.Length()
	if n.synchronized {
		lngBlend := dprec.Mix(lngFirst, lngSecond, n.factor)
		if lngFirst > lngSecond {
			factor := lngBlend / lngFirst
			return n.first.Position() * factor
		} else {
			factor := lngBlend / lngSecond
			return n.second.Position() * factor
		}
	} else {
		if lngFirst > lngSecond {
			return n.first.Position()
		} else {
			return n.second.Position()
		}
	}
}

// SetPosition sets the current position of the animation in seconds.
func (n *PairAnimationBlending) SetPosition(position float64) {
	if n.synchronized {
		lngFirst := n.first.Length()
		lngSecond := n.second.Length()
		lngBlend := dprec.Mix(lngFirst, lngSecond, n.factor)
		factor := position / lngBlend
		n.first.SetPosition(factor * lngFirst)
		n.second.SetPosition(factor * lngSecond)
	} else {
		n.first.SetPosition(position)
		n.second.SetPosition(position)
	}
}

// NodeTransform returns the transformation of the node with the specified
// name. The transformation is a blend of the transformations of the two
// sources of the node.
func (n *PairAnimationBlending) NodeTransform(name string) NodeTransform {
	switch {
	case n.factor < 0.000001: // optimization
		return n.first.NodeTransform(name)
	case n.factor > 0.999999: // optimization
		return n.second.NodeTransform(name)
	default:
		firstTransform := n.first.NodeTransform(name)
		secondTransform := n.second.NodeTransform(name)
		return BlendNodeTransforms(firstTransform, secondTransform, n.factor)
	}
}
