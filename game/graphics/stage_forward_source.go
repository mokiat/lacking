package graphics

import "github.com/mokiat/lacking/render"

func newForwardSourceStage(api render.API) *ForwardSourceStage {
	return &ForwardSourceStage{
		api: api,
	}
}

var _ Stage = (*ForwardSourceStage)(nil)

// ForwardSourceStage is a stage that provides source textures for
// a forward pass renderer.
type ForwardSourceStage struct {
	api render.API

	framebufferWidth  uint32
	framebufferHeight uint32

	hdrTexture render.Texture
}

// HDRTexture returns the texture that contains the high dynamic range
// color information.
func (s *ForwardSourceStage) HDRTexture() render.Texture {
	return s.hdrTexture
}

func (s *ForwardSourceStage) Allocate() {
	s.framebufferWidth = 32
	s.framebufferHeight = 32
	s.allocateTextures()
}

func (s *ForwardSourceStage) Release() {
	defer s.releaseTextures()
}

func (s *ForwardSourceStage) PreRender(width, height uint32) {
	if s.framebufferWidth != width || s.framebufferHeight != height {
		s.framebufferWidth = width
		s.framebufferHeight = height
		s.releaseTextures()
		s.allocateTextures()
	}
}

func (s *ForwardSourceStage) Render(ctx StageContext) {
	// Nothing to do here.
}

func (s *ForwardSourceStage) PostRender() {
	// Nothing to do here.
}

func (s *ForwardSourceStage) allocateTextures() {
	s.hdrTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           s.framebufferWidth,
		Height:          s.framebufferHeight,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
}

func (s *ForwardSourceStage) releaseTextures() {
	defer s.hdrTexture.Release()
}
