package gltfutil

import (
	"os"
	"path/filepath"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/util/blob"
	"github.com/qmuntal/gltf"
)

func RootNodeIndices(doc *gltf.Document) []uint32 {
	childrenIDs := make(map[uint32]struct{})
	for _, node := range doc.Nodes {
		for _, childID := range node.Children {
			childrenIDs[childID] = struct{}{}
		}
	}
	var result []uint32
	for id := range doc.Nodes {
		if _, ok := childrenIDs[uint32(id)]; !ok {
			result = append(result, uint32(id))
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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
	buffer := BufferViewData(doc, *accessor.BufferView)
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

func ColorTexture(doc *gltf.Document, pbr *gltf.PBRMetallicRoughness, modelURI string) []byte {
	colorTexture := pbr.BaseColorTexture
	if colorTexture == nil {
		return nil
	}
	if colorTexture.TexCoord != 0 {
		log.Warn("Unsupported color texture: tex coord layer unsupported")
	}
	texture := doc.Textures[colorTexture.Index]
	if texture.Source == nil {
		log.Warn("Unsupported color texture: no source")
		return nil
	}
	image := doc.Images[*texture.Source]
	if image.BufferView != nil {
		return BufferViewData(doc, *image.BufferView)
	} else {
		content, err := os.ReadFile(filepath.Join(filepath.Dir(modelURI), image.URI))
		if err != nil {
			log.Error("Error reading texture %q: %v", image.URI, err)
			return nil
		}
		return content
	}
}

func BufferViewData(doc *gltf.Document, index uint32) blob.Buffer {
	bufferView := doc.BufferViews[index]
	offset := bufferView.ByteOffset
	count := bufferView.ByteLength
	buffer := doc.Buffers[bufferView.Buffer]
	return blob.Buffer(buffer.Data[offset : offset+count])
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

func InverseBindMatrix(doc *gltf.Document, skin *gltf.Skin, index int) sprec.Mat4 {
	if skin.InverseBindMatrices == nil {
		log.Warn("Skin lacks inverse bind matrices")
		return sprec.IdentityMat4()
	}

	accessor := doc.Accessors[*skin.InverseBindMatrices]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return sprec.IdentityMat4()
	}

	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorMat4 {
		log.Error("Unsupported joints accessor type %d", accessor.Type)
		return sprec.IdentityMat4()
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		array := [16]float32{
			buffer.Float32(index*64 + 0*4),
			buffer.Float32(index*64 + 1*4),
			buffer.Float32(index*64 + 2*4),
			buffer.Float32(index*64 + 3*4),
			buffer.Float32(index*64 + 4*4),
			buffer.Float32(index*64 + 5*4),
			buffer.Float32(index*64 + 6*4),
			buffer.Float32(index*64 + 7*4),
			buffer.Float32(index*64 + 8*4),
			buffer.Float32(index*64 + 9*4),
			buffer.Float32(index*64 + 10*4),
			buffer.Float32(index*64 + 11*4),
			buffer.Float32(index*64 + 12*4),
			buffer.Float32(index*64 + 13*4),
			buffer.Float32(index*64 + 14*4),
			buffer.Float32(index*64 + 15*4),
		}
		return sprec.ColumnMajorArrayToMat4(array)
	default:
		log.Error("Unsupported joints accessor component type %d", accessor.ComponentType)
		return sprec.IdentityMat4()
	}
}

func AnimationKeyframes(doc *gltf.Document, sampler *gltf.AnimationSampler) []float64 {
	if sampler.Input == nil {
		log.Error("Animation sampler input is unspecified")
		return nil
	}
	accessor := doc.Accessors[*sampler.Input]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorScalar {
		log.Error("Unsupported sampler input accessor type %d", accessor.Type)
		return nil
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		result := make([]float64, accessor.Count)
		for i := 0; i < int(accessor.Count); i++ {
			result[i] = float64(buffer.Float32(i * 4))
		}
		return result
	default:
		log.Error("Unsupported sampler input accessor component type %d", accessor.ComponentType)
		return nil
	}
}

func AnimationTranslations(doc *gltf.Document, sampler *gltf.AnimationSampler) []dprec.Vec3 {
	if sampler.Output == nil {
		log.Error("Animation sampler output is unspecified")
		return nil
	}
	accessor := doc.Accessors[*sampler.Output]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorVec3 {
		log.Error("Unsupported sampler output accessor type %d", accessor.Type)
		return nil
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		result := make([]dprec.Vec3, accessor.Count)
		for i := 0; i < int(accessor.Count); i++ {
			result[i] = stod.Vec3(sprec.NewVec3(
				buffer.Float32(i*12+0*4),
				buffer.Float32(i*12+1*4),
				buffer.Float32(i*12+2*4),
			))
		}
		return result
	default:
		log.Error("Unsupported sampler output accessor component type %d", accessor.ComponentType)
		return nil
	}
}

func AnimationRotations(doc *gltf.Document, sampler *gltf.AnimationSampler) []dprec.Quat {
	if sampler.Output == nil {
		log.Error("Animation sampler output is unspecified")
		return nil
	}
	accessor := doc.Accessors[*sampler.Output]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorVec4 {
		log.Error("Unsupported sampler output accessor type %d", accessor.Type)
		return nil
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		result := make([]dprec.Quat, accessor.Count)
		for i := 0; i < int(accessor.Count); i++ {
			result[i] = stod.Quat(sprec.UnitQuat(sprec.NewQuat(
				buffer.Float32(i*16+3*4),
				buffer.Float32(i*16+0*4),
				buffer.Float32(i*16+1*4),
				buffer.Float32(i*16+2*4),
			)))
		}
		return result
	default:
		log.Error("Unsupported sampler output accessor component type %d", accessor.ComponentType)
		return nil
	}
}

func AnimationScales(doc *gltf.Document, sampler *gltf.AnimationSampler) []dprec.Vec3 {
	if sampler.Output == nil {
		log.Error("Animation sampler output is unspecified")
		return nil
	}
	accessor := doc.Accessors[*sampler.Output]
	if accessor.BufferView == nil {
		log.Warn("Accessor lacks a buffer view")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorVec3 {
		log.Error("Unsupported sampler output accessor type %d", accessor.Type)
		return nil
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		result := make([]dprec.Vec3, accessor.Count)
		for i := 0; i < int(accessor.Count); i++ {
			result[i] = stod.Vec3(sprec.NewVec3(
				buffer.Float32(i*12+0*4),
				buffer.Float32(i*12+1*4),
				buffer.Float32(i*12+2*4),
			))
		}
		return result
	default:
		log.Error("Unsupported sampler output accessor component type %d", accessor.ComponentType)
		return nil
	}
}

func IsCollisionDisabled(node *gltf.Node) bool {
	props, ok := node.Extras.(map[string]any)
	if !ok {
		return false
	}
	return props["collision"] == "none"
}
