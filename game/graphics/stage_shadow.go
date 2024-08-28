package graphics

import (
	"cmp"
	"slices"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
	"github.com/mokiat/lacking/util/spatial"
)

func newShadowStage(data *commonStageData, meshRenderer *meshRenderer) *ShadowStage {
	return &ShadowStage{
		data:         data,
		meshRenderer: meshRenderer,

		litStaticMeshes: spatial.NewVisitorBucket[uint32](65536),
		litMeshes:       spatial.NewVisitorBucket[*Mesh](1024),
	}
}

var _ Stage = (*ShadowStage)(nil)

// ShadowStage is a stage that renders shadows.
type ShadowStage struct {
	data         *commonStageData
	meshRenderer *meshRenderer

	litStaticMeshes *spatial.VisitorBucket[uint32]
	litMeshes       *spatial.VisitorBucket[*Mesh]
}

func (s *ShadowStage) Allocate() {
	// Nothing to do here.
}

func (s *ShadowStage) Release() {
	// Nothing to do here.
}

func (s *ShadowStage) PreRender(width, height uint32) {
	// Nothing to do here.
}

func (s *ShadowStage) Render(ctx StageContext) {
	defer metric.BeginRegion("shadow").End()

	s.sortLights(ctx)
	s.allocateCascadeShadowMaps(ctx)

	for _, light := range ctx.VisibleDirectionalLights {
		if !light.active { // TODO: Move this to VisitorBucket closure with iterator.
			continue
		}
		if !light.castShadow {
			continue
		}
		if light.shadowMaps == ([3]internal.CascadeShadowMapRef{}) {
			continue
		}
		s.renderDirectionalLightShadowMaps(ctx, light)
	}
}

func (s *ShadowStage) PostRender() {
	// Nothing to do here.
}

func (s *ShadowStage) sortLights(ctx StageContext) {
	slices.SortFunc(ctx.VisibleAmbientLights, func(a, b *AmbientLight) int {
		distanceA := dprec.Vec3Diff(a.Position(), ctx.CameraPosition).Length()
		distanceB := dprec.Vec3Diff(b.Position(), ctx.CameraPosition).Length()
		return cmp.Compare(distanceA, distanceB)
	})
	slices.SortFunc(ctx.VisiblePointLights, func(a, b *PointLight) int {
		distanceA := dprec.Vec3Diff(a.Position(), ctx.CameraPosition).Length()
		distanceB := dprec.Vec3Diff(b.Position(), ctx.CameraPosition).Length()
		return cmp.Compare(distanceA, distanceB)
	})
	slices.SortFunc(ctx.VisibleSpotLights, func(a, b *SpotLight) int {
		distanceA := dprec.Vec3Diff(a.Position(), ctx.CameraPosition).Length()
		distanceB := dprec.Vec3Diff(b.Position(), ctx.CameraPosition).Length()
		return cmp.Compare(distanceA, distanceB)
	})
	slices.SortFunc(ctx.VisibleDirectionalLights, func(a, b *DirectionalLight) int {
		distanceA := dprec.Vec3Diff(a.Position(), ctx.CameraPosition).Length()
		distanceB := dprec.Vec3Diff(b.Position(), ctx.CameraPosition).Length()
		return cmp.Compare(distanceA, distanceB)
	})
}

func (s *ShadowStage) allocateCascadeShadowMaps(ctx StageContext) {
	cascadeShadowMapIndex := 0

	for _, light := range ctx.VisibleDirectionalLights {
		light.shadowMaps = [3]internal.CascadeShadowMapRef{
			{
				ProjectionMatrix: sprec.IdentityMat4(),
			},
			{
				ProjectionMatrix: sprec.IdentityMat4(),
			},
			{
				ProjectionMatrix: sprec.IdentityMat4(),
			},
		}
		if !light.active { // TODO: Move this to VisitorBucket closure with iterator.
			continue
		}
		if !light.castShadow {
			continue
		}
		var shadowMaps [3]internal.CascadeShadowMapRef
		for i := range light.shadowMaps {
			if cascadeShadowMapIndex >= len(s.data.cascadeShadowMaps) {
				return
			}
			shadowMaps[i] = internal.CascadeShadowMapRef{
				CascadeShadowMap: s.data.cascadeShadowMaps[cascadeShadowMapIndex],
				ProjectionMatrix: lightOrtho(), // FIXME
			}
			cascadeShadowMapIndex++
		}
		light.shadowMaps = shadowMaps
	}
}

func (s *ShadowStage) renderDirectionalLightShadowMaps(ctx StageContext, light *DirectionalLight) {
	for _, shadowMap := range light.shadowMaps {
		shadowTexture := shadowMap.Texture
		shadowTextureWidth := float32(shadowTexture.Width())
		shadowTextureHeight := float32(shadowTexture.Height())

		projectionMatrix := shadowMap.ProjectionMatrix
		lightMatrix := light.gfxMatrix()
		lightMatrix.M14 = sprec.Floor(lightMatrix.M14*shadowTextureWidth) / shadowTextureWidth
		lightMatrix.M24 = sprec.Floor(lightMatrix.M24*shadowTextureHeight) / shadowTextureHeight
		lightMatrix.M34 = sprec.Floor(lightMatrix.M34*shadowTextureWidth) / shadowTextureWidth
		viewMatrix := sprec.InverseMat4(lightMatrix)
		projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
		frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

		s.litStaticMeshes.Reset()
		ctx.Scene.staticMeshOctree.VisitHexahedronRegion(&frustum, s.litStaticMeshes)

		s.litMeshes.Reset()
		ctx.Scene.dynamicMeshSet.VisitHexahedronRegion(&frustum, s.litMeshes)

		s.meshRenderer.DiscardRenderItems()
		for _, meshIndex := range s.litStaticMeshes.Items() {
			staticMesh := &ctx.Scene.staticMeshes[meshIndex]
			s.meshRenderer.QueueStaticMeshRenderItems(ctx, staticMesh, internal.MeshRenderPassTypeShadow)
		}
		for _, mesh := range s.litMeshes.Items() {
			s.meshRenderer.QueueMeshRenderItems(ctx, mesh, internal.MeshRenderPassTypeShadow)
		}

		commandBuffer := ctx.CommandBuffer
		uniformBuffer := ctx.UniformBuffer

		commandBuffer.BeginRenderPass(render.RenderPassInfo{
			Framebuffer: shadowMap.Framebuffer,
			Viewport: render.Area{
				Width:  shadowTexture.Width(),
				Height: shadowTexture.Height(),
			},
			DepthLoadOp:     render.LoadOperationClear,
			DepthStoreOp:    render.StoreOperationStore,
			DepthClearValue: 1.0,
			StencilLoadOp:   render.LoadOperationLoad,
			StencilStoreOp:  render.StoreOperationDiscard,
		})

		lightCameraPlacement := ubo.WriteUniform(uniformBuffer, internal.CameraUniform{
			ProjectionMatrix: projectionMatrix,
			ViewMatrix:       viewMatrix,
			CameraMatrix:     lightMatrix,
			Viewport:         sprec.ZeroVec4(), // TODO?
			Time:             ctx.Scene.Time(), // FIXME?
		})
		ctx.CameraPlacement = lightCameraPlacement // FIXME: DIRTY HACK; Use own meshRenderer context
		s.meshRenderer.Render(ctx)
		commandBuffer.EndRenderPass()
	}
}
