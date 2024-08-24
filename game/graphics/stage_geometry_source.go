package graphics

import "github.com/mokiat/lacking/render"

func newGeometrySourceStage(api render.API) *GeometrySourceStage {
	return &GeometrySourceStage{
		api: api,
	}
}

var _ Stage = (*GeometrySourceStage)(nil)

// GeometrySourceStage is a stage that provides source textures for
// a geometry pass renderer.
type GeometrySourceStage struct {
	api render.API

	framebufferWidth  uint32
	framebufferHeight uint32

	albedoMetallicTexture  render.Texture
	normalRoughnessTexture render.Texture
}

// AlbedoMetallicTexture returns the texture that contains the albedo
// color in the RGB channels and the metallic factor in the A channel.
func (s *GeometrySourceStage) AlbedoMetallicTexture() render.Texture {
	return s.albedoMetallicTexture
}

// NormalRoughnessTexture returns the texture that contains the normal
// vector in the RGB channels and the roughness factor in the A channel.
func (s *GeometrySourceStage) NormalRoughnessTexture() render.Texture {
	return s.normalRoughnessTexture
}

func (s *GeometrySourceStage) Allocate() {
	s.framebufferWidth = 32
	s.framebufferHeight = 32
	s.allocateTextures()
}

func (s *GeometrySourceStage) Release() {
	defer s.releaseTextures()
}

func (s *GeometrySourceStage) PreRender(width, height uint32) {
	if s.framebufferWidth != width || s.framebufferHeight != height {
		s.framebufferWidth = width
		s.framebufferHeight = height
		s.releaseTextures()
		s.allocateTextures()
	}
}

func (s *GeometrySourceStage) Render(ctx StageContext) {
	// Nothing to do here.
}

func (s *GeometrySourceStage) PostRender() {
	// Nothing to do here.
}

func (s *GeometrySourceStage) allocateTextures() {
	s.albedoMetallicTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           s.framebufferWidth,
		Height:          s.framebufferHeight,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
	})
	s.normalRoughnessTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           s.framebufferWidth,
		Height:          s.framebufferHeight,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
}

func (s *GeometrySourceStage) releaseTextures() {
	defer s.albedoMetallicTexture.Release()
	defer s.normalRoughnessTexture.Release()
}
