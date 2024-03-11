package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
)

// MeshDefinitionInfo contains everything needed to create a new MeshDefinition.
type MeshDefinitionInfo struct {
	Geometry  *MeshGeometry
	Materials []*Material
}

// MeshDefinition represents the definition of a mesh.
// Multiple mesh instances can be created off of one template
// reusing resources.
type MeshDefinition struct {
	engine   *Engine
	geometry *MeshGeometry

	materials      []*Material
	materialPasses [][internal.MeshRenderPassTypeCount][]internal.MeshRenderPass
	passesByType   [internal.MeshRenderPassTypeCount][]internal.MeshRenderPass
}

// MaterialCount returns the number of materials defined for this
// MeshDefinition.
func (d *MeshDefinition) MaterialCount() int {
	return len(d.materials)
}

// Material returns the material at the specified index.
func (d *MeshDefinition) Material(index int) *Material {
	return d.materials[index]
}

// SetMaterial sets the material at the specified index.
func (d *MeshDefinition) SetMaterial(index int, material *Material) {
	d.materials[index] = material
	for i := range internal.MeshRenderPassTypeCount {
		d.deleteMaterialPasses(index, internal.MeshRenderPassType(i))
		if material != nil {
			d.createMaterialPasses(index, internal.MeshRenderPassType(i))
		}
		d.updateGlobalPasses(internal.MeshRenderPassType(i))
	}
}

// Delete releases any resources owned by this MeshDefinition.
func (d *MeshDefinition) Delete() {
	for i := range len(d.materials) {
		d.SetMaterial(i, nil)
	}
	d.engine = nil
}

func (d *MeshDefinition) deleteMaterialPasses(index int, passType internal.MeshRenderPassType) {
	for _, pass := range d.materialPasses[index][passType] {
		pass.Pipeline.Release()
		pass.Program.Release()
	}
	clear(d.materialPasses[index][passType])
	d.materialPasses[index][passType] = d.materialPasses[index][passType][:0]
}

func (d *MeshDefinition) createMaterialPasses(index int, passType internal.MeshRenderPassType) {
	meshShaderInfo := internal.ShaderMeshInfo{
		VertexArray:         d.geometry.vertexArray,
		MeshHasCoords:       d.geometry.vertexFormat.HasCoord,
		MeshHasNormals:      d.geometry.vertexFormat.HasNormal,
		MeshHasTangents:     d.geometry.vertexFormat.HasTangent,
		MeshHasTextureUVs:   d.geometry.vertexFormat.HasTexCoord,
		MeshHasVertexColors: d.geometry.vertexFormat.HasColor,
		MeshHasArmature:     d.geometry.vertexFormat.HasWeights && d.geometry.vertexFormat.HasJoints,
	}

	fragment := d.geometry.fragments[index]
	material := d.materials[index]

	switch passType {
	case internal.MeshRenderPassTypeGeometry:
		for _, pass := range material.geometryPasses {
			programCode := pass.Shader.CreateProgramCode(internal.GeometryShaderProgramCodeInfo{
				ShaderMeshInfo: meshShaderInfo,
			})
			program := d.engine.createGeometryPassProgram(programCode)
			pipeline := d.engine.createGeometryPassPipeline(internal.RenderPassPipelineInfo{
				Program:          program,
				MeshVertexArray:  d.geometry.vertexArray,
				FragmentTopology: fragment.topology,
				PassDefinition:   pass.MaterialRenderPassDefinition,
			})
			d.materialPasses[index][passType] = append(d.materialPasses[index][passType], internal.MeshRenderPass{
				MeshRenderPassDefinition: internal.MeshRenderPassDefinition{
					Layer:           pass.Layer,
					Program:         program,
					Pipeline:        pipeline,
					IndexByteOffset: fragment.indexByteOffset,
					IndexCount:      fragment.indexCount,
				},
				Key:          d.engine.pickFreeRenderPassKey(),
				Textures:     pass.Textures,
				Samplers:     pass.Samplers,
				MaterialData: pass.UniformData,
			})
		}

	case internal.MeshRenderPassTypeShadow:
		for _, pass := range material.shadowPasses {
			programCode := pass.Shader.CreateProgramCode(internal.ShadowShaderProgramCodeInfo{
				ShaderMeshInfo: meshShaderInfo,
			})
			program := d.engine.createShadowPassProgram(programCode)
			pipeline := d.engine.createShadowPassPipeline(internal.RenderPassPipelineInfo{
				Program:          program,
				MeshVertexArray:  d.geometry.vertexArray,
				FragmentTopology: fragment.topology,
				PassDefinition:   pass.MaterialRenderPassDefinition,
			})
			d.materialPasses[index][passType] = append(d.materialPasses[index][passType], internal.MeshRenderPass{
				MeshRenderPassDefinition: internal.MeshRenderPassDefinition{
					Layer:           pass.Layer,
					Program:         program,
					Pipeline:        pipeline,
					IndexByteOffset: fragment.indexByteOffset,
					IndexCount:      fragment.indexCount,
				},
				Key:          d.engine.pickFreeRenderPassKey(),
				Textures:     pass.Textures,
				Samplers:     pass.Samplers,
				MaterialData: pass.UniformData,
			})
		}

	case internal.MeshRenderPassTypeForward:
		for _, pass := range material.forwardPasses {
			programCode := pass.Shader.CreateProgramCode(internal.ForwardShaderProgramCodeInfo{
				ShaderMeshInfo: meshShaderInfo,
			})
			program := d.engine.createForwardPassProgram(programCode)
			pipeline := d.engine.createForwardPassPipeline(internal.RenderPassPipelineInfo{
				Program:          program,
				MeshVertexArray:  d.geometry.vertexArray,
				FragmentTopology: fragment.topology,
				PassDefinition:   pass.MaterialRenderPassDefinition,
			})
			d.materialPasses[index][passType] = append(d.materialPasses[index][passType], internal.MeshRenderPass{
				MeshRenderPassDefinition: internal.MeshRenderPassDefinition{
					Layer:           pass.Layer,
					Program:         program,
					Pipeline:        pipeline,
					IndexByteOffset: fragment.indexByteOffset,
					IndexCount:      fragment.indexCount,
				},
				Key:          d.engine.pickFreeRenderPassKey(),
				Textures:     pass.Textures,
				Samplers:     pass.Samplers,
				MaterialData: pass.UniformData,
			})
		}
	}
}

func (d *MeshDefinition) updateGlobalPasses(passType internal.MeshRenderPassType) {
	clear(d.passesByType[passType])
	d.passesByType[passType] = d.passesByType[passType][:0]
	for _, passes := range d.materialPasses {
		d.passesByType[passType] = append(d.passesByType[passType], passes[passType]...)
	}
}
