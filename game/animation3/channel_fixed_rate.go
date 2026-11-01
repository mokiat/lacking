package animation3

// FixedRateKeyframeChannel is a [Channel] that keeps a sequence of complete
// transformations sampled at a fixed rate, starting at a given timestamp.
// Frame i is located at startTimestamp + i / frameRate, so frames do not need
// to store their own timestamps and looking up a frame does not require a
// search.
//
// Two consecutive frames are interpolated in the same way as
// [BlendTransforms] blends two transformations.
type FixedRateKeyframeChannel struct {
	startTimestamp float64
	frameRate      float64
	frames         []Transform
}

var _ Channel = (*FixedRateKeyframeChannel)(nil)

// NewFixedRateKeyframeChannel creates a new [FixedRateKeyframeChannel] whose
// first frame is located at startTimestamp and whose frames follow each other
// at frameRate frames per second of keyframe time.
//
// The frame rate must be positive and there must be at least one frame; the
// constructor panics otherwise. A single frame produces a channel that holds
// a constant transformation.
//
// The channel takes ownership of the frames slice; the caller must not
// modify it afterwards.
func NewFixedRateKeyframeChannel(startTimestamp, frameRate float64, frames []Transform) *FixedRateKeyframeChannel {
	if !(frameRate > 0.0) {
		panic("frame rate must be positive")
	}
	if len(frames) == 0 {
		panic("at least one frame is required")
	}
	return &FixedRateKeyframeChannel{
		startTimestamp: startTimestamp,
		frameRate:      frameRate,
		frames:         frames,
	}
}

// Transform returns the transformation at the specified keyframe timestamp,
// interpolated between the two frames that surround the timestamp. A
// timestamp before the first or after the last frame holds the value of that
// frame.
func (c *FixedRateKeyframeChannel) Transform(timestamp float64) Transform {
	position := (timestamp - c.startTimestamp) * c.frameRate
	if position <= 0.0 {
		return c.frames[0]
	}
	lastIndex := len(c.frames) - 1
	if position >= float64(lastIndex) {
		return c.frames[lastIndex]
	}
	index := int(position)
	fraction := position - float64(index)
	return BlendTransforms(c.frames[index], c.frames[index+1], fraction)
}
