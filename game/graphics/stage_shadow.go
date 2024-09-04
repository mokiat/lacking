package graphics

import (
	"cmp"
	"slices"

	"github.com/mokiat/gog"
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
	s.distributeCascadeShadowMaps(ctx)

	for _, light := range ctx.VisibleDirectionalLights {
		if !light.active { // TODO: Move this to VisitorBucket closure with iterator.
			continue
		}
		if !light.castShadow {
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

func (s *ShadowStage) distributeCascadeShadowMaps(ctx StageContext) {
	s.data.ResetDirectionalShadowMapAssignments()

	for _, light := range ctx.VisibleDirectionalLights {
		if !light.active { // TODO: Move this to VisitorBucket closure with iterator.
			continue
		}
		if !light.castShadow {
			continue
		}

		shadowMap := s.data.AssignDirectionalShadowMap(light)
		if shadowMap == nil {
			return // no more free shadow maps
		}

		lightMatrix := light.gfxMatrix()
		lightXAxis := lightMatrix.OrientationX()
		lightYAxis := lightMatrix.OrientationY()
		lightZAxis := lightMatrix.OrientationZ()
		shadowView := sprec.InverseMat4(sprec.OrientationMat4(lightXAxis, lightYAxis, lightZAxis))

		cascadeCount := min(len(ctx.Camera.cascadeDistances), len(shadowMap.Cascades))
		gog.MutateIndex(shadowMap.Cascades[:cascadeCount], func(j int, cascade *internal.DirectionalShadowMapCascade) {
			cascadeNear := ctx.Camera.CascadeNear(j)
			cascadeFar := ctx.Camera.CascadeFar(j)
			cameraProjectionMatrix := cascadeProjectionMatrix(ctx, cascadeNear, cascadeFar)

			cameraModelMatrix := ctx.Camera.gfxMatrix()
			cameraViewMatrix := sprec.InverseMat4(cameraModelMatrix)
			cameraInverseProjectionView := sprec.InverseMat4(
				sprec.Mat4Prod(cameraProjectionMatrix, cameraViewMatrix),
			)

			frustumCornerPoints4D := [8]sprec.Vec4{
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(-1, -1, -1, 1.0)),
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(1, -1, -1, 1.0)),
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(-1, 1, -1, 1.0)),
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(1, 1, -1, 1.0)),
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(-1, -1, 1, 1.0)),
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(1, -1, 1, 1.0)),
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(-1, 1, 1, 1.0)),
				sprec.Mat4Vec4Prod(cameraInverseProjectionView, sprec.NewVec4(1, 1, 1, 1.0)),
			}
			var frustumCornerPoints [8]sprec.Vec3
			for i, point := range frustumCornerPoints4D {
				frustumCornerPoints[i] = sprec.Vec3Quot(point.VecXYZ(), point.W)
			}

			// TODO: make these configurable or based on camera frustum
			shadowNearOverflow := float32(100.0)
			shadowFarOverflow := float32(100.0)

			shadowNear := -(maxDirection(lightZAxis, frustumCornerPoints) + shadowNearOverflow)
			shadowFar := -(minDirection(lightZAxis, frustumCornerPoints) - shadowFarOverflow)
			shadowLeft := minDirection(lightXAxis, frustumCornerPoints)
			shadowRight := maxDirection(lightXAxis, frustumCornerPoints)
			shadowBottom := minDirection(lightYAxis, frustumCornerPoints)
			shadowTop := maxDirection(lightYAxis, frustumCornerPoints)
			shadowOrtho := sprec.OrthoMat4(shadowLeft, shadowRight, shadowTop, shadowBottom, shadowNear, shadowFar)
			shadowMatrix := sprec.Mat4Prod(shadowOrtho, shadowView)

			cascade.Near = cascadeNear
			cascade.Far = cascadeFar
			cascade.ProjectionMatrix = shadowMatrix
		})
	}
}

func (s *ShadowStage) renderDirectionalLightShadowMaps(ctx StageContext, light *DirectionalLight) {
	shadowMap := s.data.GetDirectionalShadowMap(light)
	if shadowMap == nil {
		return
	}

	cascadeCount := min(len(ctx.Camera.cascadeDistances), len(shadowMap.Cascades))
	for _, cascade := range shadowMap.Cascades[:cascadeCount] {
		frustum := spatial.ProjectionRegion(stod.Mat4(cascade.ProjectionMatrix))

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
		shadowTexture := shadowMap.ArrayTexture

		commandBuffer.BeginRenderPass(render.RenderPassInfo{
			Framebuffer: cascade.Framebuffer,
			Viewport: render.Area{
				Width:  shadowTexture.Width(),
				Height: shadowTexture.Height(),
			},
			DepthLoadOp:     render.LoadOperationClear,
			DepthStoreOp:    render.StoreOperationStore,
			DepthClearValue: 1.0,
			// DepthBias:       64 * 1024.0,
			DepthSlopeBias: 2.0,
			StencilLoadOp:  render.LoadOperationLoad,
			StencilStoreOp: render.StoreOperationDiscard,
		})
		lightCameraPlacement := ubo.WriteUniform(uniformBuffer, internal.CameraUniform{
			ProjectionMatrix: cascade.ProjectionMatrix,
			ViewMatrix:       sprec.IdentityMat4(), // irrelevant
			CameraMatrix:     sprec.IdentityMat4(), // irrelevant
			Viewport:         sprec.ZeroVec4(),     // TODO?
			Time:             ctx.Scene.Time(),     // FIXME?
		})
		ctx.CameraPlacement = lightCameraPlacement // FIXME: DIRTY HACK; Use own meshRenderer context
		s.meshRenderer.Render(ctx)
		commandBuffer.EndRenderPass()
	}
}

func cascadeProjectionMatrix(ctx StageContext, near, far float32) sprec.Mat4 {
	oldNear := ctx.Camera.near
	oldFar := ctx.Camera.far

	ctx.Camera.near = near
	ctx.Camera.far = far

	cameraProjectionMatrix := evaluateProjectionMatrix(ctx.Camera, ctx.Viewport.Width, ctx.Viewport.Height)

	ctx.Camera.near = oldNear
	ctx.Camera.far = oldFar

	return cameraProjectionMatrix
}

func minDirection(direction sprec.Vec3, points [8]sprec.Vec3) float32 {
	return min(
		sprec.Vec3Dot(direction, points[0]),
		sprec.Vec3Dot(direction, points[1]),
		sprec.Vec3Dot(direction, points[2]),
		sprec.Vec3Dot(direction, points[3]),
		sprec.Vec3Dot(direction, points[4]),
		sprec.Vec3Dot(direction, points[5]),
		sprec.Vec3Dot(direction, points[6]),
		sprec.Vec3Dot(direction, points[7]),
	)
}

func maxDirection(direction sprec.Vec3, points [8]sprec.Vec3) float32 {
	return max(
		sprec.Vec3Dot(direction, points[0]),
		sprec.Vec3Dot(direction, points[1]),
		sprec.Vec3Dot(direction, points[2]),
		sprec.Vec3Dot(direction, points[3]),
		sprec.Vec3Dot(direction, points[4]),
		sprec.Vec3Dot(direction, points[5]),
		sprec.Vec3Dot(direction, points[6]),
		sprec.Vec3Dot(direction, points[7]),
	)
}
