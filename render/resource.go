package render

// Resource represents an object that consumes resources and needs
// to be released when no longer in use.
type Resource interface {

	// Release releases any resources allocated by this object.
	Release()
}
