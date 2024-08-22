package graphics

import (
	"math"
	"slices"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render/ubo"
)

const (
	initialRenderItemCount = 32 * 1024

	// TODO: Move these next to the uniform types
	modelUniformBufferItemSize  = 64
	modelUniformBufferItemCount = 256
	modelUniformBufferSize      = modelUniformBufferItemSize * modelUniformBufferItemCount
)

func newMeshRenderer() *meshRenderer {
	return &meshRenderer{
		renderItems:            make([]renderItem, 0, initialRenderItemCount),
		modelUniformBufferData: make(gblob.LittleEndianBlock, modelUniformBufferSize),
	}
}

type meshRenderer struct {
	renderItems            []renderItem
	modelUniformBufferData gblob.LittleEndianBlock
}

func (s *meshRenderer) DiscardRenderItems() {
	s.renderItems = s.renderItems[:0]
}

func (s *meshRenderer) QueueMeshRenderItems(ctx StageContext, mesh *Mesh, passType internal.MeshRenderPassType) {
	if !mesh.active {
		return
	}
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

			IndexByteOffset: pass.IndexByteOffset,
			IndexCount:      pass.IndexCount,
		})
	}
}

func (s *meshRenderer) QueueStaticMeshRenderItems(ctx StageContext, mesh *StaticMesh, passType internal.MeshRenderPassType) {
	if !mesh.active {
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
			ArmatureKey: math.MaxUint32,

			Pipeline:     pass.Pipeline,
			TextureSet:   pass.TextureSet,
			UniformSet:   pass.UniformSet,
			ModelData:    mesh.matrixData,
			ArmatureData: nil,

			IndexByteOffset: pass.IndexByteOffset,
			IndexCount:      pass.IndexCount,
		})
	}
}

func (s *meshRenderer) Render(ctx StageContext) {
	s.renderMeshRenderItems(ctx, s.renderItems)
	s.renderItems = s.renderItems[:0]
}

func (s *meshRenderer) renderMeshRenderItems(ctx StageContext, items []renderItem) {
	const maxBatchSize = modelUniformBufferItemCount
	var (
		lastMaterialKey = uint32(math.MaxUint32)
		lastArmatureKey = uint32(math.MaxUint32)

		batchStart = 0
		batchEnd   = 0
	)

	slices.SortFunc(items, compareMeshRenderItems)

	itemCount := len(items)
	for i, item := range items {
		materialKey := item.MaterialKey
		armatureKey := item.ArmatureKey

		isSame := (materialKey == lastMaterialKey) && (armatureKey == lastArmatureKey)
		if !isSame {
			if batchStart < batchEnd {
				s.renderMeshRenderItemBatch(ctx, items[batchStart:batchEnd])
			}
			batchStart = batchEnd
		}
		batchEnd++

		batchSize := batchEnd - batchStart
		if (batchSize >= maxBatchSize) || (i == itemCount-1) {
			s.renderMeshRenderItemBatch(ctx, items[batchStart:batchEnd])
			batchStart = batchEnd
		}

		lastMaterialKey = materialKey
		lastArmatureKey = armatureKey
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
	for i, item := range items {
		start := i * modelUniformBufferItemSize
		end := start + modelUniformBufferItemSize
		copy(s.modelUniformBufferData[start:end], item.ModelData)
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
