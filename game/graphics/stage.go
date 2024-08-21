package graphics

import "github.com/mokiat/lacking/render"

// Stage represents a render stage (e.g. geometry, lighting, post-processing).
type Stage interface {

	// Allocate is called once in the beginning to initialize any graphics
	// resources.
	Allocate()

	// Release is called once in the end to release any graphics resources.
	Release()

	// Resize is called whenever the screen is resized. The width and height
	// are in pixels and will never be zero or less.
	//
	// The implementation may choose to ignore this call if it does not need
	// to adjust its resources.
	Resize(width, height uint32)

	// Render is called whenever the stage should render its content.
	Render(ctx StageContext)
}

// StageContext represents the context that is passed to a render stage.
type StageContext struct {

	// Camera is the camera that should be used to render the stage.
	Camera *Camera

	// Viewport is the area of the screen that the stage should render to.
	// The width and height of the viewport will match the width and height
	// that were passed to the last Resize method call.
	Viewport render.Area

	// Framebuffer is the screen framebuffer. A stage would not normally use
	// this unless it is the last stage in the rendering pipeline.
	Framebuffer render.Framebuffer
}

// StageTextureParameter is a function that returns a texture that is used as
// a parameter to a render stage.
type StageTextureParameter func() render.Texture
