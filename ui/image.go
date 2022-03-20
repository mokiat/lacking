package ui

// Image represents a 2D image.
type Image interface {

	// Size returns the dimensions of this Image.
	Size() Size

	// Destroy releases all resources allocated for this
	// image.
	Destroy()
}
