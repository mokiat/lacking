package animation3

// Channel represents the animation data of a single bone in a [Clip]. A
// Channel always animates its bone; a bone that is not animated by a clip
// has no Channel in it.
//
// Implementations are free to choose how they store and interpolate the
// animation data, but a Channel is expected to be immutable once created, so
// that it can be shared by any number of consumers.
type Channel interface {

	// Transform returns the local transformation of the bone, relative to its
	// parent, at the specified keyframe timestamp. The returned value is
	// always a complete transformation, including any components that the
	// underlying animation data does not vary. A timestamp outside the range
	// covered by the animation data holds the value at the nearest end of
	// that range.
	Transform(timestamp float64) Transform
}
