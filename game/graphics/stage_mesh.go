package graphics

import (
	"cmp"
	"sort"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog/seq"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render/ubo"
)

const (
	initialRenderItemCount = 32 * 1024

	// TODO: Move these next to the uniform types
	modelUniformBufferItemSize  = 64 // 1x mat4
	modelUniformBufferItemCount = 256
	modelUniformBufferSize      = modelUniformBufferItemSize * modelUniformBufferItemCount

	instanceUniformBufferItemSize  = 16 // 1x vec4
	instanceUniformBufferItemCount = 256
	instanceUniformBufferSize      = instanceUniformBufferItemSize * instanceUniformBufferItemCount
)

func newMeshRenderer() *meshRenderer {
	return &meshRenderer{
		renderItems:               make([]renderItem, 0, initialRenderItemCount),
		modelUniformBufferData:    make(gblob.LittleEndianBlock, modelUniformBufferSize),
		instanceUniformBufferData: make(gblob.LittleEndianBlock, instanceUniformBufferSize),
	}
}

type meshRenderer struct {
	renderItems               []renderItem
	modelUniformBufferData    gblob.LittleEndianBlock
	instanceUniformBufferData gblob.LittleEndianBlock
}

func (s *meshRenderer) DiscardRenderItems() {
	s.renderItems = s.renderItems[:0]
}

func (s *meshRenderer) QueueMeshRenderItems(ctx StageContext, mesh *Mesh, passType internal.MeshRenderPassType) {
	if !mesh.active {
		return
	}
	if ctx.Cascade > mesh.maxCascade {
		return
	}
	definition := mesh.definition
	passes := definition.passesByType[passType]
	for _, pass := range passes {
		s.renderItems = append(s.renderItems, renderItem{
			Layer:       pass.Layer,
			MaterialKey: pass.Key,
			ArmatureKey: mesh.armature.key(),

			Pipeline:   pass.Pipeline,
			TextureSet: pass.TextureSet,
			UniformSet: pass.UniformSet,

			ModelData:    mesh.matrixData,
			ArmatureData: mesh.armature.uniformData(),
			InstanceData: mesh.instanceData,

			IndexByteOffset: pass.IndexByteOffset,
			IndexCount:      pass.IndexCount,
		})
	}
}

func (s *meshRenderer) QueueStaticMeshRenderItems(ctx StageContext, mesh *StaticMesh, passType internal.MeshRenderPassType) {
	if !mesh.active {
		return
	}
	if ctx.Cascade > mesh.maxCascade {
		return
	}
	distance := dprec.Vec3Diff(mesh.position, ctx.CameraPosition).Length()
	if distance < mesh.minDistance || mesh.maxDistance < distance {
		return
	}

	// TODO: Extract common stuff between mesh and static mesh into a type
	// that is passed ot this function instead so that it can be reused.
	definition := mesh.definition
	passes := definition.passesByType[passType]
	for _, pass := range passes {
		s.renderItems = append(s.renderItems, renderItem{
			Layer:       pass.Layer,
			MaterialKey: pass.Key,
			ArmatureKey: mesh.armature.key(),

			Pipeline:     pass.Pipeline,
			TextureSet:   pass.TextureSet,
			UniformSet:   pass.UniformSet,
			ModelData:    mesh.matrixData,
			ArmatureData: mesh.armature.uniformData(),
			InstanceData: mesh.instanceData,

			IndexByteOffset: pass.IndexByteOffset,
			IndexCount:      pass.IndexCount,
		})
	}
}

func (s *meshRenderer) Render(ctx StageContext) {
	s.sortRenderItems(s.renderItems)
	s.renderMeshRenderItems(ctx, s.renderItems)
	s.renderItems = s.renderItems[:0]
}

func (s *meshRenderer) sortRenderItems(items []renderItem) {
	sort.Slice(items, func(i, j int) bool {
		a, b := &items[i], &items[j]
		return cmp.Or(
			cmp.Compare(a.Layer, b.Layer),
			cmp.Compare(a.MaterialKey, b.MaterialKey),
			cmp.Compare(a.ArmatureKey, b.ArmatureKey),
		) == -1
	})
}

func (s *meshRenderer) renderMeshRenderItems(ctx StageContext, items []renderItem) {
	iter := seq.BatchSliceFast(items, itemEqFunc, modelUniformBufferItemCount)

	for batch := range iter {
		s.renderMeshRenderItemBatch(ctx, batch)
	}
}

func (s *meshRenderer) renderMeshRenderItemBatch(ctx StageContext, items []renderItem) {
	template := items[0]

	commandBuffer := ctx.CommandBuffer
	uniformBuffer := ctx.UniformBuffer

	commandBuffer.BindPipeline(template.Pipeline)

	// Camera data is shared between all items.
	cameraPlacement := ctx.CameraPlacement
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		cameraPlacement.Buffer,
		cameraPlacement.Offset,
		cameraPlacement.Size,
	)

	// Material data is shared between all items.
	if !template.UniformSet.IsEmpty() {
		materialPlacement := ubo.WriteUniform(uniformBuffer, internal.MaterialUniform{
			Data: template.UniformSet.Data(),
		})
		commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingMaterial,
			materialPlacement.Buffer,
			materialPlacement.Offset,
			materialPlacement.Size,
		)
	}

	for i := range template.TextureSet.TextureCount() {
		if texture := template.TextureSet.TextureAt(i); texture != nil {
			commandBuffer.TextureUnit(uint(i), texture)
		}
		if sampler := template.TextureSet.SamplerAt(i); sampler != nil {
			commandBuffer.SamplerUnit(uint(i), sampler)
		}
	}

	// Model data needs to be combined.
	for i := range items {
		item := &items[i]

		modelStart := i * modelUniformBufferItemSize
		modelEnd := modelStart + modelUniformBufferItemSize
		copy(s.modelUniformBufferData[modelStart:modelEnd], item.ModelData)

		instanceStart := i * instanceUniformBufferItemSize
		instanceEnd := instanceStart + instanceUniformBufferItemSize
		copy(s.instanceUniformBufferData[instanceStart:instanceEnd], item.InstanceData[:])
	}
	modelPlacement := ubo.WriteUniform(uniformBuffer, internal.ModelUniform{
		ModelMatrices: s.modelUniformBufferData,
	})
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingModel,
		modelPlacement.Buffer,
		modelPlacement.Offset,
		modelPlacement.Size,
	)
	instancePlacement := ubo.WriteUniform(uniformBuffer, internal.InstanceUniform{
		InstanceBlocks: s.instanceUniformBufferData,
	})
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingTiming,
		instancePlacement.Buffer,
		instancePlacement.Offset,
		instancePlacement.Size,
	)

	// Armature data is shared between all items.
	if template.ArmatureData != nil {
		armaturePlacement := ubo.WriteUniform(uniformBuffer, internal.ArmatureUniform{
			BoneMatrices: template.ArmatureData,
		})
		commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingArmature,
			armaturePlacement.Buffer,
			armaturePlacement.Offset,
			armaturePlacement.Size,
		)
	}

	commandBuffer.DrawIndexed(template.IndexByteOffset, template.IndexCount, uint32(len(items)))
}

func itemEqFunc(items []renderItem, i, j int) bool {
	a := &items[i]
	b := &items[j]
	return a.MaterialKey == b.MaterialKey && a.ArmatureKey == b.ArmatureKey
}
