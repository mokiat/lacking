package animation

// Source represents a source of animation data.
type Source interface {

	// Length returns the length of the animation in seconds.
	Length() float64

	// Position returns the current position of the animation in seconds.
	Position() float64

	// SetPosition sets the current position of the animation in seconds.
	SetPosition(position float64)

	// NodeTransform returns the transformation of the node with the
	// specified name at the current time position.
	NodeTransform(name string) NodeTransform
}
