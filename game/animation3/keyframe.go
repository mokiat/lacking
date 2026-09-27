package animation3

import "github.com/mokiat/gomath/dprec"

// Keyframe represents a single keyframe in an animation.
type Keyframe[T any] struct {
	Timestamp float64
	Value     T
}

// KeyframeList is a list of keyframes ordered by ascending timestamp.
type KeyframeList[T any] []Keyframe[T]

// Sample returns the two keyframes that surround the specified timestamp and
// the interpolation factor between them.
//
// A timestamp outside the covered range clamps to the nearest keyframe. An
// empty list returns zero-value keyframes and a factor of 0.0.
func (l KeyframeList[T]) Sample(timestamp float64) (Keyframe[T], Keyframe[T], float64) {
	switch len(l) {
	case 0:
		var zero Keyframe[T]
		return zero, zero, 0.0
	case 1:
		return l[0], l[0], 0.0
	default:
		leftIndex, rightIndex := l.binarySearch(timestamp)
		left := l[leftIndex]
		right := l[rightIndex]
		if leftIndex == rightIndex {
			return left, right, 0.0
		}
		if right.Timestamp <= left.Timestamp {
			return left, right, 0.0
		}
		t := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
		return left, right, dprec.Clamp(t, 0.0, 1.0)
	}
}

func (l KeyframeList[T]) binarySearch(timestamp float64) (int, int) {
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
	return leftIndex, rightIndex
}
