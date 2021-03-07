package command

type ChangeFramebuffer struct {
	Framebuffer OptionalFramebuffer
	Viewport    OptionalArea
	Scissor     OptionalArea
}
