package graphics

import "github.com/mokiat/lacking/render"

func newDepthSourceStage(api render.API) *DepthSourceStage {
	return &DepthSourceStage{
		api: api,
	}
}

var _ Stage = (*DepthSourceStage)(nil)

// DepthSourceStage is a stage that provides a depth source texture.
type DepthSourceStage struct {
	api render.API

	framebufferWidth  uint32
	framebufferHeight uint32

	depthTexture render.Texture
}

// DepthTexture returns the texture that contains the depth information.
func (s *DepthSourceStage) DepthTexture() render.Texture {
	return s.depthTexture
}

func (s *DepthSourceStage) Allocate() {
	s.framebufferWidth = 32
	s.framebufferHeight = 32
	s.allocateTextures()
}

func (s *DepthSourceStage) Release() {
	defer s.releaseTextures()
}

func (s *DepthSourceStage) PreRender(width, height uint32) {
	if s.framebufferWidth != width || s.framebufferHeight != height {
		s.framebufferWidth = width
		s.framebufferHeight = height
		s.releaseTextures()
		s.allocateTextures()
	}
}

func (s *DepthSourceStage) Render(ctx StageContext) {
	// Nothing to do here.
}

func (s *DepthSourceStage) PostRender() {
	// Nothing to do here.
}

func (s *DepthSourceStage) allocateTextures() {
	s.depthTexture = s.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Label:  "Depth Source Texture",
		Width:  s.framebufferWidth,
		Height: s.framebufferHeight,
	})
}

func (s *DepthSourceStage) releaseTextures() {
	defer s.depthTexture.Release()
}
