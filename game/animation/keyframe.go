package animation

import "github.com/mokiat/gomath/dprec"

// Keyframe represents a single keyframe in an animation.
type Keyframe[T any] struct {
	Timestamp float64
	Value     T
}

// KeyframeList is a list of keyframes.
type KeyframeList[T any] []Keyframe[T]

// Keyframe returns the keyframes that are closest to the specified timestamp
// and the interpolation factor between them.
func (l KeyframeList[T]) Keyframe(timestamp float64) (Keyframe[T], Keyframe[T], float64) {
	leftIndex := 0
	rightIndex := len(l) - 1
	for leftIndex < rightIndex-1 {
		middleIndex := (leftIndex + rightIndex) / 2
		middle := l[middleIndex]
		if middle.Timestamp <= timestamp {
			leftIndex = middleIndex
		}
		if middle.Timestamp >= timestamp {
			rightIndex = middleIndex
		}
	}
	left := l[leftIndex]
	right := l[rightIndex]
	if leftIndex == rightIndex {
		return left, right, 0
	}
	t := dprec.Clamp((timestamp-left.Timestamp)/(right.Timestamp-left.Timestamp), 0.0, 1.0)
	return left, right, t
}

// KeyframeSet represents a set of keyframes for an animation, including
// translations, rotations, and scales.
type KeyframeSet struct {
	TranslationKeyframes KeyframeList[dprec.Vec3]
	RotationKeyframes    KeyframeList[dprec.Quat]
	ScaleKeyframes       KeyframeList[dprec.Vec3]
}

// Translation returns the translation at the specified timestamp by
// interpolating between the two closest keyframes.
func (s KeyframeSet) Translation(timestamp float64) dprec.Vec3 {
	left, right, t := s.TranslationKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

// Rotation returns the rotation at the specified timestamp by interpolating
// between the two closest keyframes using spherical linear interpolation.
func (s KeyframeSet) Rotation(timestamp float64) dprec.Quat {
	left, right, t := s.RotationKeyframes.Keyframe(timestamp)
	return dprec.QuatSlerp(left.Value, right.Value, t)
}

// Scale returns the scale at the specified timestamp by interpolating between
// the two closest keyframes.
func (s KeyframeSet) Scale(timestamp float64) dprec.Vec3 {
	left, right, t := s.ScaleKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}
