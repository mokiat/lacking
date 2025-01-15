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
			cameraProjectionViewMatrix := sprec.Mat4Prod(cameraProjectionMatrix, cameraViewMatrix)

			frustumCornerPoints := calculateFrustumCornerPoints(cameraProjectionViewMatrix)
			frustumCentralPoint := calculateFrustumCentralPoint(frustumCornerPoints)
			frustumRadius := sprec.Ceil(calculateFrustumRadius(frustumCentralPoint, frustumCornerPoints))

			frustumOffsetX := sprec.Vec3Dot(lightXAxis, frustumCentralPoint)
			frustumOffsetY := sprec.Vec3Dot(lightYAxis, frustumCentralPoint)
			frustumOffsetZ := sprec.Vec3Dot(lightZAxis, frustumCentralPoint)

			// TODO: make these configurable or based on visible objects in
			// view frustum.
			shadowNearOverflow := float32(200.0)
			shadowFarOverflow := float32(100.0)

			shadowNear := -frustumOffsetZ - (frustumRadius + shadowNearOverflow)
			shadowFar := -frustumOffsetZ + (frustumRadius + shadowFarOverflow)
			shadowLeft := frustumOffsetX - frustumRadius
			shadowRight := frustumOffsetX + frustumRadius
			shadowBottom := frustumOffsetY - frustumRadius
			shadowTop := frustumOffsetY + frustumRadius

			shadowOrtho := sprec.OrthoMat4(shadowLeft, shadowRight, shadowTop, shadowBottom, shadowNear, shadowFar)
			shadowMatrix := sprec.Mat4Prod(shadowOrtho, shadowView)

			// NOTE: If this ever has precision problems, consider moving the
			// anchor point on fixed intervals. If that also fails, consider
			// using double precision for the error calculation.
			anchorPointVec4 := sprec.NewVec4(0.0, 0.0, 0.0, 1.0)
			projectedAnchorPoint4 := sprec.Mat4Vec4Prod(shadowMatrix, anchorPointVec4)
			projectedAnchorPoint := sprec.Vec3Quot(projectedAnchorPoint4.VecXYZ(), projectedAnchorPoint4.W)

			shadowMapWidth := float32(shadowMap.ArrayTexture.Width())
			shadowMapHeight := float32(shadowMap.ArrayTexture.Height())

			projectedAnchorPoint.X = (projectedAnchorPoint.X + 1.0) * 0.5 * shadowMapWidth
			projectedAnchorPoint.Y = (projectedAnchorPoint.Y + 1.0) * 0.5 * shadowMapHeight

			diffX := projectedAnchorPoint.X - sprec.Floor(projectedAnchorPoint.X)
			diffY := projectedAnchorPoint.Y - sprec.Floor(projectedAnchorPoint.Y)

			errorX := diffX * (shadowRight - shadowLeft) / shadowMapWidth
			errorY := diffY * (shadowTop - shadowBottom) / shadowMapHeight

			shadowLeft += errorX
			shadowRight += errorX
			shadowBottom += errorY
			shadowTop += errorY

			shadowOrtho = sprec.OrthoMat4(shadowLeft, shadowRight, shadowTop, shadowBottom, shadowNear, shadowFar)
			shadowMatrix = sprec.Mat4Prod(shadowOrtho, shadowView)

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
	for i, cascade := range shadowMap.Cascades[:cascadeCount] {
		frustum := spatial.ProjectionRegion(stod.Mat4(cascade.ProjectionMatrix))

		s.litStaticMeshes.Reset()
		ctx.Scene.staticMeshOctree.VisitHexahedronRegion(&frustum, s.litStaticMeshes)

		s.litMeshes.Reset()
		ctx.Scene.dynamicMeshSet.VisitHexahedronRegion(&frustum, s.litMeshes)

		ctx.Cascade = uint8(i + 1)

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

func calculateFrustumCornerPoints(projectionMatrix sprec.Mat4) [8]sprec.Vec3 {
	inverseProjectionMatrix := sprec.InverseMat4(projectionMatrix)
	frustumCornerPoints4D := [8]sprec.Vec4{
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(-1, -1, -1, 1.0)),
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(1, -1, -1, 1.0)),
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(-1, 1, -1, 1.0)),
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(1, 1, -1, 1.0)),
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(-1, -1, 1, 1.0)),
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(1, -1, 1, 1.0)),
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(-1, 1, 1, 1.0)),
		sprec.Mat4Vec4Prod(inverseProjectionMatrix, sprec.NewVec4(1, 1, 1, 1.0)),
	}
	var frustumCornerPoints [8]sprec.Vec3
	for i, point := range frustumCornerPoints4D {
		frustumCornerPoints[i] = sprec.Vec3Quot(point.VecXYZ(), point.W)
	}
	return frustumCornerPoints
}

func calculateFrustumCentralPoint(cornerPoints [8]sprec.Vec3) sprec.Vec3 {
	var result sprec.Vec3
	for _, point := range cornerPoints {
		result = sprec.Vec3Sum(result, point)
	}
	return sprec.Vec3Quot(result, 8.0)
}

func calculateFrustumRadius(centralPoint sprec.Vec3, cornerPoints [8]sprec.Vec3) float32 {
	var result float32
	for _, point := range cornerPoints {
		distance := sprec.Vec3Diff(point, centralPoint).Length()
		result = max(result, distance)
	}
	return result
}
