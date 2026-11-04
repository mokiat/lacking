package animation

import "github.com/mokiat/gomath/dprec"

// TransformKeyframeChannel is a [Channel] that keeps a single list of
// keyframes, each holding a complete transformation, so that the translation,
// rotation and scale components all share the same timestamps.
//
// Two consecutive keyframes are interpolated in the same way as
// [BlendTransforms] blends two transformations.
type TransformKeyframeChannel struct {
	frames KeyframeList[Transform]
}

var _ Channel = (*TransformKeyframeChannel)(nil)

// NewTransformKeyframeChannel creates a new [TransformKeyframeChannel] from
// the specified keyframes. The list must contain at least one keyframe and
// the constructor panics if it is empty. A single keyframe produces a channel
// that holds a constant transformation.
//
// The channel takes ownership of the keyframe list; the caller must not
// modify it afterwards.
func NewTransformKeyframeChannel(frames KeyframeList[Transform]) *TransformKeyframeChannel {
	if len(frames) == 0 {
		panic("at least one keyframe is required")
	}
	return &TransformKeyframeChannel{
		frames: frames,
	}
}

// Transform returns the transformation at the specified keyframe timestamp,
// interpolated between the two keyframes that surround the timestamp. A
// timestamp before the first or after the last keyframe holds the value of
// that keyframe.
func (c *TransformKeyframeChannel) Transform(timestamp float64) Transform {
	leftIndex, rightIndex := c.frames.binarySearch(timestamp)
	if leftIndex == rightIndex {
		return c.frames[leftIndex].Value
	}
	left := c.frames[leftIndex]
	right := c.frames[rightIndex]
	a := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
	return BlendTransforms(left.Value, right.Value, dprec.Clamp(a, 0.0, 1.0))
}
