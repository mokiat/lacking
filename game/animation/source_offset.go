package animation

// NewOffsetSource creates a new instance of OffsetSource. An OffsetSource
// can be used to align two looping animations. For non-looping animations,
// the length returned will not be adjusted.
func NewOffsetSource(delegate Source) *OffsetSource {
	return &OffsetSource{
		delegate: delegate,
		offset:   0.0,
	}
}

// OffsetSource is a decorator for an animation source that allows one to
// shift the timestamp, making it appear as though the animation starts
// earlier or later than normal.
type OffsetSource struct {
	delegate Source
	offset   float64
}

// Source returns the underlying animation source.
func (s *OffsetSource) Source() Source {
	return s.delegate
}

// Offset returns the timestamp offset of the animation. A positive value
// means that the animation starts later (i.e. shifted to the right)
func (s *OffsetSource) Offset() float64 {
	return s.offset
}

// SetOffset changes the offset of the animation.
func (s *OffsetSource) SetOffset(offset float64) *OffsetSource {
	s.offset = offset
	return s
}

// Length returns the length of the animation in seconds.
func (s *OffsetSource) Length() float64 {
	return s.delegate.Length()
}

// BoneTransformDelta returns the transformation that occurred for the
// specified bone between the from and to animation timestamps .
//
// This is mostly used for root motion.
func (s *OffsetSource) BoneTransformDelta(name string, fromTimestamp, toTimestamp float64) NodeTransform {
	return s.delegate.BoneTransformDelta(name, fromTimestamp-s.offset, toTimestamp-s.offset)
}

// BoneTransformAt returns the transformation for the specified bone
// at the specified animation timestamp.
func (s *OffsetSource) BoneTransformAt(name string, timestamp float64) NodeTransform {
	return s.delegate.BoneTransformAt(name, timestamp-s.offset)
}

// WithOffset is a helper function that creates a new OffsetSource with
// the specified offset.
func WithOffset(source Source, offset float64) *OffsetSource {
	result := NewOffsetSource(source)
	result.SetOffset(offset)
	return result
}
