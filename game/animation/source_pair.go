package animation

import "github.com/mokiat/gomath/dprec"

// NewPairBlendSource creates a new pair animation blending node
// with the specified sources.
func NewPairBlendSource(first, second Source) *PairBlendSource {
	return &PairBlendSource{
		first:        first,
		second:       second,
		synchronized: true,
	}
}

var _ Source = (*PairBlendSource)(nil)

// PairBlendSource represents an animation source that blends two
// animation sources. The blending factor is determined by the factor
// field of the node.
type PairBlendSource struct {
	first        Source
	second       Source
	synchronized bool
	factor       float64
}

// First returns the first source of the node.
func (s *PairBlendSource) First() Source {
	return s.first
}

// SetFirst sets the first source of the node.
func (s *PairBlendSource) SetFirst(first Source) {
	s.first = first
}

// Second returns the second source of the node.
func (s *PairBlendSource) Second() Source {
	return s.second
}

// SetSecond sets the second source of the node.
func (s *PairBlendSource) SetSecond(second Source) {
	s.second = second
}

// Synchronized returns whether the two sources of the node are synchronized.
func (s *PairBlendSource) Synchronized() bool {
	return s.synchronized
}

// SetSynchronized sets whether the two sources of the node are synchronized.
func (s *PairBlendSource) SetSynchronized(synchronized bool) {
	s.synchronized = synchronized
}

// Factor returns the blending factor of the node. A value of 0.0 means
// that the first source is used, a value of 1.0 means that the second
// source is used. The value is clamped to the range [0.0, 1.0].
func (s *PairBlendSource) Factor() float64 {
	return s.factor
}

// SetFactor sets the blending factor of the node. The value is clamped
// to the range [0.0, 1.0].
func (s *PairBlendSource) SetFactor(factor float64) {
	s.factor = dprec.Clamp(factor, 0.0, 1.0)
}

// Length returns the length of the animation in seconds.
func (s *PairBlendSource) Length() float64 {
	lngFirst := s.first.Length()
	lngSecond := s.second.Length()
	if s.synchronized {
		return dprec.Mix(lngFirst, lngSecond, s.factor)
	} else {
		return max(lngFirst, lngSecond)
	}
}

// Position returns the current position of the animation in seconds.
func (s *PairBlendSource) Position() float64 {
	lngFirst := s.first.Length()
	lngSecond := s.second.Length()
	if s.synchronized {
		lngBlend := dprec.Mix(lngFirst, lngSecond, s.factor)
		if lngFirst > lngSecond {
			factor := lngBlend / lngFirst
			return s.first.Position() * factor
		} else {
			factor := lngBlend / lngSecond
			return s.second.Position() * factor
		}
	} else {
		if lngFirst > lngSecond {
			return s.first.Position()
		} else {
			return s.second.Position()
		}
	}
}

// SetPosition sets the current position of the animation in seconds.
func (s *PairBlendSource) SetPosition(position float64) {
	if s.synchronized {
		lngFirst := s.first.Length()
		lngSecond := s.second.Length()
		lngBlend := dprec.Mix(lngFirst, lngSecond, s.factor)
		factor := position / lngBlend
		s.first.SetPosition(factor * lngFirst)
		s.second.SetPosition(factor * lngSecond)
	} else {
		s.first.SetPosition(position)
		s.second.SetPosition(position)
	}
}

// NodeTransform returns the transformation of the node with the specified
// name. The transformation is a blend of the transformations of the two
// sources of the node.
func (s *PairBlendSource) NodeTransform(name string) NodeTransform {
	switch {
	case s.factor < 0.000001: // optimization
		return s.first.NodeTransform(name)
	case s.factor > 0.999999: // optimization
		return s.second.NodeTransform(name)
	default:
		firstTransform := s.first.NodeTransform(name)
		secondTransform := s.second.NodeTransform(name)
		return BlendNodeTransforms(firstTransform, secondTransform, s.factor)
	}
}
