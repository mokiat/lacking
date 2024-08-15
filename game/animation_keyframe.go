package game

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
