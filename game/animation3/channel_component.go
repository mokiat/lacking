package animation3

import "github.com/mokiat/gomath/dprec"

// ComponentKeyframeChannel is a [Channel] that keeps a separate list of
// keyframes for each of the translation, rotation and scale components, so
// that each component can be keyed at its own timestamps.
//
// Translation and scale are interpolated linearly. Rotation is interpolated
// through normalized linear interpolation along the shorter arc, which does
// not keep a constant angular velocity between keyframes and is only
// noticeably different from spherical interpolation when keyframes are far
// apart in rotation.
type ComponentKeyframeChannel struct {
	translations KeyframeList[dprec.Vec3]
	rotations    KeyframeList[dprec.Quat]
	scales       KeyframeList[dprec.Vec3]
}

var _ Channel = (*ComponentKeyframeChannel)(nil)

// NewComponentKeyframeChannel creates a new [ComponentKeyframeChannel] from the
// specified keyframes. Each list must contain at least one keyframe, so a
// component that is not animated needs a single keyframe that holds its
// constant value, for example the rest value of the bone. The constructor
// panics if any of the lists is empty.
//
// The channel takes ownership of the keyframe lists; the caller must not
// modify them afterwards.
func NewComponentKeyframeChannel(
	translations KeyframeList[dprec.Vec3],
	rotations KeyframeList[dprec.Quat],
	scales KeyframeList[dprec.Vec3],
) *ComponentKeyframeChannel {
	if len(translations) == 0 {
		panic("at least one translation keyframe is required")
	}
	if len(rotations) == 0 {
		panic("at least one rotation keyframe is required")
	}
	if len(scales) == 0 {
		panic("at least one scale keyframe is required")
	}

	return &ComponentKeyframeChannel{
		translations: translations,
		rotations:    rotations,
		scales:       scales,
	}
}

// Transform returns the transformation at the specified keyframe timestamp.
// Each component is interpolated between the two of its keyframes that
// surround the timestamp. A timestamp before the first or after the last
// keyframe of a component holds the value of that keyframe.
func (c *ComponentKeyframeChannel) Transform(timestamp float64) Transform {
	return Transform{
		Translation: c.translationAt(timestamp),
		Rotation:    c.rotationAt(timestamp),
		Scale:       c.scaleAt(timestamp),
	}
}

// translationAt returns the interpolated translation at the timestamp.
func (c *ComponentKeyframeChannel) translationAt(timestamp float64) dprec.Vec3 {
	leftIndex, rightIndex := c.translations.binarySearch(timestamp)
	if leftIndex == rightIndex {
		return c.translations[leftIndex].Value
	}
	left := c.translations[leftIndex]
	right := c.translations[rightIndex]
	a := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, dprec.Clamp(a, 0.0, 1.0))
}

// rotationAt returns the interpolated rotation at the timestamp.
func (c *ComponentKeyframeChannel) rotationAt(timestamp float64) dprec.Quat {
	leftIndex, rightIndex := c.rotations.binarySearch(timestamp)
	if leftIndex == rightIndex {
		return c.rotations[leftIndex].Value
	}
	left := c.rotations[leftIndex]
	right := c.rotations[rightIndex]
	a := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
	return dprec.QuatNLerp(left.Value, right.Value, dprec.Clamp(a, 0.0, 1.0))
}

// scaleAt returns the interpolated scale at the timestamp.
func (c *ComponentKeyframeChannel) scaleAt(timestamp float64) dprec.Vec3 {
	leftIndex, rightIndex := c.scales.binarySearch(timestamp)
	if leftIndex == rightIndex {
		return c.scales[leftIndex].Value
	}
	left := c.scales[leftIndex]
	right := c.scales[rightIndex]
	a := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, dprec.Clamp(a, 0.0, 1.0))
}
