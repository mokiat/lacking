package graphics

import (
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

// GeometryStageInput is used to configure a new GeometryStage.
type GeometryStageInput struct {
	AlbedoMetallicTexture  StageTextureParameter
	NormalRoughnessTexture StageTextureParameter
	DepthTexture           StageTextureParameter
}

func newGeometryStage(api render.API, meshRenderer *meshRenderer, input GeometryStageInput) *GeometryStage {
	return &GeometryStage{
		api:          api,
		meshRenderer: meshRenderer,
		input:        input,
	}
}

var _ Stage = (*GeometryStage)(nil)

// GeometryStage is a render stage that renders the geometry of the scene.
type GeometryStage struct {
	api          render.API
	meshRenderer *meshRenderer
	input        GeometryStageInput

	albedoMetallicTexture  render.Texture
	normalRoughnessTexture render.Texture
	depthTexture           render.Texture
	framebuffer            render.Framebuffer
}

func (s *GeometryStage) Allocate() {
	s.allocateFramebuffer()
}

func (s *GeometryStage) Release() {
	defer s.releaseFramebuffer()
}

func (s *GeometryStage) PreRender(width, height uint32) {
	albedoMetallicTexture := s.input.AlbedoMetallicTexture()
	normalRoughnessTexture := s.input.NormalRoughnessTexture()
	depthTexture := s.input.DepthTexture()
	if albedoMetallicTexture != s.albedoMetallicTexture || normalRoughnessTexture != s.normalRoughnessTexture || depthTexture != s.depthTexture {
		s.releaseFramebuffer()
		s.allocateFramebuffer()
	}
}

func (s *GeometryStage) Render(ctx StageContext) {
	defer metric.BeginRegion("geometry").End()

	commandBuffer := ctx.CommandBuffer
	commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: s.framebuffer,
		Viewport: render.Area{
			// TODO: This should be based on HDR texture size.
			Width:  ctx.Viewport.Width,
			Height: ctx.Viewport.Height,
		},
		DepthLoadOp:     render.LoadOperationClear,
		DepthStoreOp:    render.StoreOperationStore,
		DepthClearValue: 1.0,
		StencilLoadOp:   render.LoadOperationLoad,
		StencilStoreOp:  render.StoreOperationDiscard,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 1.0},
			},
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 1.0, 0.0},
			},
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 1.0},
			},
		},
	})
	s.meshRenderer.DiscardRenderItems()
	for _, mesh := range ctx.VisibleMeshes {
		s.meshRenderer.QueueMeshRenderItems(ctx, mesh, internal.MeshRenderPassTypeGeometry)
	}
	for _, meshIndex := range ctx.VisibleStaticMeshIndices {
		staticMesh := &ctx.Scene.staticMeshes[meshIndex]
		s.meshRenderer.QueueStaticMeshRenderItems(ctx, staticMesh, internal.MeshRenderPassTypeGeometry)
	}
	s.meshRenderer.Render(ctx)
	commandBuffer.EndRenderPass()
}

func (s *GeometryStage) PostRender() {
	// Nothing to do here.
}

func (s *GeometryStage) allocateFramebuffer() {
	s.albedoMetallicTexture = s.input.AlbedoMetallicTexture()
	s.normalRoughnessTexture = s.input.NormalRoughnessTexture()
	s.depthTexture = s.input.DepthTexture()

	s.framebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			s.albedoMetallicTexture,
			s.normalRoughnessTexture,
		},
		DepthAttachment: s.depthTexture,
	})
}

func (s *GeometryStage) releaseFramebuffer() {
	defer s.framebuffer.Release()
}
