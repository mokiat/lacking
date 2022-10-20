package shape

// Shape represent an arbitrary 3D shape.
type Shape interface {

	// BoundingSphereRadius returns the distance from the center of the shape
	// after which the shape is no longer present.
	//
	// Ideally this radius should be the smallest possible to get the best
	// performance optimization.
	BoundingSphereRadius() float64
}
