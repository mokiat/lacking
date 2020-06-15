package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/data/gltf"
)

type GLTFProvider interface {
	GLTF() (GLTFDocument, error)
}

type GLTFResourceFile struct {
	Resource
}

func (f *GLTFResourceFile) GLTF() (GLTFDocument, error) {
	doc, err := gltf.Parse(gltf.NewFileSource(f.filename))
	if err != nil {
		return GLTFDocument{}, fmt.Errorf("failed to parse gltf model: %w", err)
	}
	return GLTFDocument{doc}, nil
}

type GLTFDocument struct {
	*gltf.Document
}

func (d GLTFDocument) FindMesh(index int) GLTFMesh {
	return GLTFMesh{
		doc:  d,
		Mesh: d.Meshes[index],
	}
}

type GLTFMesh struct {
	doc GLTFDocument
	gltf.Mesh
}

func (m GLTFMesh) FindPrimitive(index int) GLTFPrimitive {
	return GLTFPrimitive{
		doc:       m.doc,
		Primitive: m.Primitives[index],
	}
}

type GLTFPrimitive struct {
	doc GLTFDocument
	gltf.Primitive
}

func (p GLTFPrimitive) FindMode() int {
	if p.Mode == nil {
		return gltf.ModeTriangles
	}
	return *p.Mode
}

func (p GLTFPrimitive) FindMaterial() GLTFMaterial {
	if p.Material == nil {
		panic("primitive lacks material")
	}
	return GLTFMaterial{
		doc:      p.doc,
		Material: p.doc.Materials[*p.Material],
	}
}

func (p GLTFPrimitive) FindIndexCount() int {
	if p.Indices == nil {
		panic("missing indices: unsupported")
	}
	accessor := p.doc.Accessors[*p.Indices]
	return accessor.Count
}

func (p GLTFPrimitive) FindIndex(index int) int {
	if p.Indices == nil {
		return index
	}
	accessor := p.doc.Accessors[*p.Indices]
	if accessor.BufferView == nil {
		return index
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	switch accessor.ComponentType {
	case gltf.ComponentTypeUnsignedShort:
		return int(buffer.UInt16(2 * index))
	default:
		panic(fmt.Errorf("unsupported index component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindCoord(index int) [3]float32 {
	if !p.HasAttribute(gltf.AttributePosition) {
		return [3]float32{}
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributePosition]]
	if accessor.BufferView == nil {
		return [3]float32{}
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec3 {
		panic(fmt.Errorf("unsupported coord type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
		return [3]float32{
			buffer.Float32(3*4*index + 4*0),
			buffer.Float32(3*4*index + 4*1),
			buffer.Float32(3*4*index + 4*2),
		}
	default:
		panic(fmt.Errorf("unsupported coord component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindNormal(index int) [3]float32 {
	if !p.HasAttribute(gltf.AttributeNormal) {
		return [3]float32{}
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeNormal]]
	if accessor.BufferView == nil {
		return [3]float32{}
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec3 {
		panic(fmt.Errorf("unsupported normal type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
		return [3]float32{
			buffer.Float32(3*4*index + 4*0),
			buffer.Float32(3*4*index + 4*1),
			buffer.Float32(3*4*index + 4*2),
		}
	default:
		panic(fmt.Errorf("unsupported normal component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindTangent(index int) [3]float32 {
	if !p.HasAttribute(gltf.AttributeTangent) {
		return [3]float32{}
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeTangent]]
	if accessor.BufferView == nil {
		return [3]float32{}
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec3 {
		panic(fmt.Errorf("unsupported tangent type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
		return [3]float32{
			buffer.Float32(3*4*index + 4*0),
			buffer.Float32(3*4*index + 4*1),
			buffer.Float32(3*4*index + 4*2),
		}
	default:
		panic(fmt.Errorf("unsupported tangent component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindTexCoord0(index int) [2]float32 {
	if !p.HasAttribute(gltf.AttributeTexCoord0) {
		return [2]float32{}
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeTexCoord0]]
	if accessor.BufferView == nil {
		return [2]float32{}
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec2 {
		panic(fmt.Errorf("unsupported tex coord type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
		return [2]float32{
			buffer.Float32(2*4*index + 4*0),
			buffer.Float32(2*4*index + 4*1),
		}
	default:
		panic(fmt.Errorf("unsupported tex coord component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindColor0(index int) [4]float32 {
	if !p.HasAttribute(gltf.AttributeColor0) {
		return [4]float32{}
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeColor0]]
	if accessor.BufferView == nil {
		return [4]float32{}
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec4 {
		panic(fmt.Errorf("unsupported color type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
		return [4]float32{
			buffer.Float32(4*4*index + 4*0),
			buffer.Float32(4*4*index + 4*1),
			buffer.Float32(4*4*index + 4*2),
			buffer.Float32(4*4*index + 4*3),
		}
	default:
		panic(fmt.Errorf("unsupported color component type %d", accessor.ComponentType))
	}
}

type GLTFMaterial struct {
	doc GLTFDocument
	gltf.Material
}

func (m GLTFMaterial) FindMetallic() float32 {
	if m.PBRMetallicRoughness == nil {
		panic("material lacks pbr metallic roughness")
	}
	if m.PBRMetallicRoughness.MetallicFactor == nil {
		return 1.0
	}
	return *m.PBRMetallicRoughness.MetallicFactor
}

func (m GLTFMaterial) FindRoughness() float32 {
	if m.PBRMetallicRoughness == nil {
		panic("material lacks pbr metallic roughness")
	}
	if m.PBRMetallicRoughness.RoughnessFactor == nil {
		return 1.0
	}
	return *m.PBRMetallicRoughness.RoughnessFactor
}

func (m GLTFMaterial) FindBaseColor() (gltf.Color, bool) {
	if m.PBRMetallicRoughness == nil {
		return [4]float32{}, false
	}
	if m.PBRMetallicRoughness.BaseColorFactor == nil {
		return [4]float32{}, false
	}
	return *m.PBRMetallicRoughness.BaseColorFactor, true
}

func (m GLTFMaterial) FindColorTexture() (string, bool) {
	if m.PBRMetallicRoughness == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.BaseColorTexture == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.BaseColorTexture.TexCoord > 0 {
		panic("mesh material uses multiple uv coords")
	}
	texture := m.doc.Textures[m.PBRMetallicRoughness.BaseColorTexture.Index]
	if texture.Source == nil {
		panic("texture lacks a source")
	}
	image := m.doc.Images[*texture.Source]
	return image.Name, true
}

func (m GLTFMaterial) FindRoughnessTexture() (string, bool) {
	if m.PBRMetallicRoughness == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.MetallicRoughnessTexture == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.MetallicRoughnessTexture.TexCoord > 0 {
		panic("mesh material uses multiple uv coords")
	}
	texture := m.doc.Textures[m.PBRMetallicRoughness.MetallicRoughnessTexture.Index]
	if texture.Source == nil {
		panic("texture lacks a source")
	}
	image := m.doc.Images[*texture.Source]
	return image.Name, true
}

func (m GLTFMaterial) FindNormalTexture() (string, float32, bool) {
	if m.NormalTexture == nil {
		return "", 0.0, false
	}
	if m.NormalTexture.TexCoord > 0 {
		panic("mesh material uses multiple uv coords")
	}
	texture := m.doc.Textures[m.NormalTexture.Index]
	if texture.Source == nil {
		panic("texture lacks a source")
	}
	image := m.doc.Images[*texture.Source]
	scale := float32(1.0)
	if m.NormalTexture.Scale != nil {
		scale = *m.NormalTexture.Scale
	}
	return image.Name, scale, true
}
