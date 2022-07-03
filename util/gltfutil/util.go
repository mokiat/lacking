package gltfutil

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/log"
	"github.com/qmuntal/gltf"
)

func RootNodes(doc *gltf.Document) []*gltf.Node {
	childrenIDs := make(map[uint32]struct{})
	for _, node := range doc.Nodes {
		for _, childID := range node.Children {
			childrenIDs[childID] = struct{}{}
		}
	}
	var result []*gltf.Node
	for id, node := range doc.Nodes {
		if _, ok := childrenIDs[uint32(id)]; !ok {
			result = append(result, node)
		}
	}
	return result
}

func IndexCount(doc *gltf.Document, primitive *gltf.Primitive) int {
	if primitive.Indices == nil {
		log.Warn("Primitive uses no indices")
		return 0
	}
	accessor := doc.Accessors[*primitive.Indices]
	return int(accessor.Count)
}

func HasAttribute(primitive *gltf.Primitive, name string) bool {
	_, ok := primitive.Attributes[name]
	return ok
}

func Index(doc *gltf.Document, primitive *gltf.Primitive, at int) int {
	accessor := doc.Accessors[*primitive.Indices]
	if accessor.BufferView == nil {
		log.Error("Accessor lacks a buffer view")
		return 0
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	switch accessor.ComponentType {
	case gltf.ComponentUbyte:
		return int(buffer[at])
	case gltf.ComponentUshort:
		return int(buffer.Uint16(2 * at))
	case gltf.ComponentUint:
		return int(buffer.Uint32(4 * at))
	default:
		log.Error("Unsupported index accessor component type %d", accessor.ComponentType)
		return 0
	}
}

func Coord(doc *gltf.Document, primitive *gltf.Primitive, at int) sprec.Vec3 {
	if !HasAttribute(primitive, gltf.POSITION) {
		return sprec.ZeroVec3()
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.POSITION]]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return sprec.ZeroVec3()
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec3 {
		log.Error("Unsupported coord accessor type %d", accessor.Type)
		return sprec.ZeroVec3()
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec3(
			buffer.Float32(3*4*at+4*0),
			buffer.Float32(3*4*at+4*1),
			buffer.Float32(3*4*at+4*2),
		)
	default:
		log.Error("Unsupported coord accessor component type %d", accessor.ComponentType)
		return sprec.ZeroVec3()
	}
}

func Normal(doc *gltf.Document, primitive *gltf.Primitive, at int) sprec.Vec3 {
	if !HasAttribute(primitive, gltf.NORMAL) {
		return sprec.ZeroVec3()
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.NORMAL]]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return sprec.ZeroVec3()
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec3 {
		log.Error("Unsupported normal accessor type %d", accessor.Type)
		return sprec.ZeroVec3()
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec3(
			buffer.Float32(3*4*at+4*0),
			buffer.Float32(3*4*at+4*1),
			buffer.Float32(3*4*at+4*2),
		)
	default:
		log.Error("Unsupported normal accessor component type %d", accessor.ComponentType)
		return sprec.ZeroVec3()
	}
}

func Tangent(doc *gltf.Document, primitive *gltf.Primitive, at int) sprec.Vec3 {
	if !HasAttribute(primitive, gltf.TANGENT) {
		return sprec.ZeroVec3()
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.TANGENT]]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return sprec.ZeroVec3()
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec3 {
		log.Error("Unsupported tangent accessor type %d", accessor.Type)
		return sprec.ZeroVec3()
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec3(
			buffer.Float32(3*4*at+4*0),
			buffer.Float32(3*4*at+4*1),
			buffer.Float32(3*4*at+4*2),
		)
	default:
		log.Error("Unsupported tangent accessor component type %d", accessor.ComponentType)
		return sprec.ZeroVec3()
	}
}

func TexCoord0(doc *gltf.Document, primitive *gltf.Primitive, at int) sprec.Vec2 {
	if !HasAttribute(primitive, gltf.TEXCOORD_0) {
		return sprec.ZeroVec2()
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.TEXCOORD_0]]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return sprec.ZeroVec2()
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec2 {
		log.Error("Unsupported tex coord accessor type %d", accessor.Type)
		return sprec.ZeroVec2()
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec2(
			buffer.Float32(2*4*at+4*0),
			1.0-buffer.Float32(2*4*at+4*1), // fix tex coord orientation
		)
	default:
		log.Error("Unsupported tex coord accessor component type %d", accessor.ComponentType)
		return sprec.ZeroVec2()
	}
}

func Color0(doc *gltf.Document, primitive *gltf.Primitive, at int) sprec.Vec4 {
	if !HasAttribute(primitive, gltf.COLOR_0) {
		return sprec.ZeroVec4()
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.COLOR_0]]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return sprec.ZeroVec4()
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec4 {
		log.Error("Unsupported color accessor type %d", accessor.Type)
		return sprec.ZeroVec4()
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec4(
			buffer.Float32(4*4*at+4*0),
			buffer.Float32(4*4*at+4*1),
			buffer.Float32(4*4*at+4*2),
			buffer.Float32(4*4*at+4*3),
		)
	default:
		log.Error("Unsupported color accessor component type %d", accessor.ComponentType)
		return sprec.ZeroVec4()
	}
}

func Weights0(doc *gltf.Document, primitive *gltf.Primitive, at int) sprec.Vec4 {
	if !HasAttribute(primitive, gltf.WEIGHTS_0) {
		return sprec.ZeroVec4()
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.WEIGHTS_0]]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return sprec.ZeroVec4()
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec4 {
		log.Error("Unsupported weights accessor type %d", accessor.Type)
		return sprec.ZeroVec4()
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec4(
			buffer.Float32(4*4*at+4*0),
			buffer.Float32(4*4*at+4*1),
			buffer.Float32(4*4*at+4*2),
			buffer.Float32(4*4*at+4*3),
		)
	default:
		log.Error("Unsupported weights accessor component type %d", accessor.ComponentType)
		return sprec.ZeroVec4()
	}
}

func Joints0(doc *gltf.Document, primitive *gltf.Primitive, at int) [4]uint8 {
	if !HasAttribute(primitive, gltf.JOINTS_0) {
		return [4]uint8{}
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.JOINTS_0]]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return [4]uint8{}
	}
	bufferView := doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec4 {
		log.Error("Unsupported joints accessor type %d", accessor.Type)
		return [4]uint8{}
	}
	switch accessor.ComponentType {
	case gltf.ComponentUbyte:
		return [4]uint8{
			buffer.Uint8(4*at + 0),
			buffer.Uint8(4*at + 1),
			buffer.Uint8(4*at + 2),
			buffer.Uint8(4*at + 3),
		}
	case gltf.ComponentUshort:
		return [4]uint8{
			uint8(buffer.Uint16(4*2*at + 0*2)),
			uint8(buffer.Uint16(4*2*at + 1*2)),
			uint8(buffer.Uint16(4*2*at + 2*2)),
			uint8(buffer.Uint16(4*2*at + 3*2)),
		}
	default:
		log.Error("Unsupported joints accessor component type %d", accessor.ComponentType)
		return [4]uint8{}
	}
}

func PrimitiveMaterial(doc *gltf.Document, primitive *gltf.Primitive) *gltf.Material {
	if primitive.Material == nil {
		return nil
	}
	return doc.Materials[*primitive.Material]
}

func BaseColor(pbr *gltf.PBRMetallicRoughness) sprec.Vec4 {
	factor := pbr.BaseColorFactorOrDefault()
	return sprec.NewVec4(factor[0], factor[1], factor[2], factor[3])
}

func ColorTexture(doc *gltf.Document, pbr *gltf.PBRMetallicRoughness) string {
	colorTexture := pbr.BaseColorTexture
	if colorTexture == nil {
		return ""
	}
	if colorTexture.TexCoord != 0 {
		log.Warn("Unsupported color texture: tex coord layer unsupported")
	}
	texture := doc.Textures[colorTexture.Index]
	if texture.Source == nil {
		log.Warn("Unsupported color texture: no source")
		return ""
	}
	image := doc.Images[*texture.Source]
	return image.Name
}

func MetallicRoughnessTexture(doc *gltf.Document, pbr *gltf.PBRMetallicRoughness) string {
	mrTexture := pbr.MetallicRoughnessTexture
	if mrTexture == nil {
		return ""
	}
	if mrTexture.TexCoord != 0 {
		log.Warn("Unsupported metallic-roughness texture: tex coord layer unsupported")
	}
	texture := doc.Textures[mrTexture.Index]
	if texture.Source == nil {
		log.Warn("Unsupported metallic-roughness texture: no source")
		return ""
	}
	image := doc.Images[*texture.Source]
	return image.Name
}

func NormalTexture(doc *gltf.Document, material *gltf.Material) (string, float32) {
	normalTexture := material.NormalTexture
	if normalTexture == nil {
		return "", 1.0
	}
	if normalTexture.TexCoord > 0 {
		log.Warn("Unsupported normal texture: tex coord layer unsupported")
	}
	if normalTexture.Index == nil {
		log.Error("Normal texture lacks an index")
		return "", normalTexture.ScaleOrDefault()
	}
	texture := doc.Textures[*normalTexture.Index]
	if texture.Source == nil {
		log.Warn("Unsupported normal texture: no source")
		return "", normalTexture.ScaleOrDefault()
	}
	image := doc.Images[*texture.Source]
	return image.Name, normalTexture.ScaleOrDefault()
}