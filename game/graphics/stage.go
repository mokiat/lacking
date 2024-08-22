package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

// Stage represents a render stage (e.g. geometry, lighting, post-processing).
type Stage interface {

	// Allocate is called once in the beginning to initialize any graphics
	// resources.
	Allocate()

	// Release is called once in the end to release any graphics resources.
	Release()

	// PreRender is called before the stage renders its content. The width and
	// height are in pixels and will never be zero or less.
	PreRender(width, height uint32)

	// Render is called whenever the stage should render its content.
	Render(ctx StageContext)

	// PostRender is called after all commands have been queued to the render API.
	PostRender()
}

// StageContext represents the context that is passed to a render stage.
type StageContext struct {

	// Scene is the scene that should be rendered by the stage.
	Scene *Scene

	// Camera is the camera that should be used to render the stage.
	Camera *Camera

	// CameraPosition is the position of the camera in world space.
	CameraPosition dprec.Vec3

	// CameraPlacement is the uniform buffer segment that contains the camera
	// data.
	CameraPlacement ubo.UniformPlacement

	// VisibleMeshes is a list of meshes that are visible in the scene.
	VisibleMeshes []*Mesh

	// VisibleStaticMeshIndices is a list of indices of static meshes that are
	// visible in the scene.
	VisibleStaticMeshIndices []uint32

	// DebugLines is a list of debug lines that should be rendered by the stage.
	//
	// FIXME: Figure out a different way to do this. Maybe make it easy for
	// users to emulate debug lines through forward pass and mutatable meshes?
	DebugLines []DebugLine

	// Viewport is the area of the screen that the stage should render to.
	// The width and height of the viewport will match the width and height
	// that were passed to the PreRender method call.
	Viewport render.Area

	// Framebuffer is the screen framebuffer. A stage would not normally use
	// this unless it is the last stage in the rendering pipeline.
	Framebuffer render.Framebuffer

	// CommandBuffer is the command buffer that the stage should use to queue
	// rendering commands.
	CommandBuffer render.CommandBuffer

	// UniformBuffer is the uniform buffer that the stage should use to set
	// uniform data.
	UniformBuffer *ubo.UniformBlockBuffer
}

// StageTextureParameter is a function that returns a texture that is used as
// a parameter to a render stage.
type StageTextureParameter func() render.Texture
