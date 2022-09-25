package shape

// Shape represent a 3D shape.
type Shape interface {

	// BoundingSphereRadius returns the radius from the center of the shape
	// after which the shape is no longer present.
	//
	// Ideally this radius should be the smallest possible to get the most
	// performance optimization.
	BoundingSphereRadius() float64
}
