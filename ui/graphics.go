package ui

// func newGraphics(renderAPI render.API, shaders ShaderCollection) Graphics {
// 	return nil // TODO
// }

// type

// // Graphics is an interface through which the framework
// // can issue drawing calls.
// type Graphics interface {

// 	// Create initializes the graphics API.
// 	Create()

// 	// Resize indicates a new draw area size to the
// 	// graphics API.
// 	Resize(size Size)

// 	// ResizeFramebuffer indicates a new framebuffer size
// 	// to the graphics API. This can be useful when the size
// 	// and the resolution don't match (higher dpi).
// 	ResizeFramebuffer(size Size)

// 	// CreateImage creates a new Image that can be used
// 	// in draw operations.
// 	CreateImage(img image.Image) (Image, error)

// 	// ReleaseImage releases any resources allocated for the
// 	// given Image resource.
// 	ReleaseImage(resource Image) error

// 	// CreateFront creates a new Font that can be used
// 	// in text draw operations.
// 	CreateFont(font *opentype.Font) (Font, error)

// 	// ReleaseFont releases any resources allocated for the
// 	// given Font resource.
// 	ReleaseFont(font Font) error

// 	// Begin prepares the graphics API for drawing.
// 	Begin()

// 	// Canvas returns a Canvas instance that can be
// 	// used to issue various draw operations.
// 	Canvas() Canvas

// 	// End flushes any accumulated state in the graphics
// 	// API to the screen.
// 	End()

// 	// Destroy releases any resources allocated by the
// 	// graphics API.
// 	Destroy()
// }
