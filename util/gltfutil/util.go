package gltfutil

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
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
	result := make([]uint32, 0, len(doc.Nodes)-len(childrenIDs))
	for id := range doc.Nodes {
		if _, ok := childrenIDs[uint32(id)]; !ok {
			result = append(result, uint32(id))
		}
	}
	return result
}

func HasAttribute(primitive *gltf.Primitive, name string) bool {
	_, ok := primitive.Attributes[name]
	return ok
}

func Indices(doc *gltf.Document, primitive *gltf.Primitive) ([]int, error) {
	if primitive.Indices == nil {
		return nil, nil
	}
	accessor := doc.Accessors[*primitive.Indices]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([]int, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentUbyte:
		for i := range result {
			result[i] = int(scanner.ScanUint8())
		}
	case gltf.ComponentUshort:
		for i := range result {
			result[i] = int(scanner.ScanUint16())
		}
	case gltf.ComponentUint:
		for i := range result {
			result[i] = int(scanner.ScanUint32())
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func Coords(doc *gltf.Document, primitive *gltf.Primitive) ([]sprec.Vec3, error) {
	if !HasAttribute(primitive, gltf.POSITION) {
		return nil, nil
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.POSITION]]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	if accessor.Type != gltf.AccessorVec3 {
		return nil, fmt.Errorf("unsupported accessor type %d", accessor.Type)
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([]sprec.Vec3, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		for i := range result {
			result[i] = scanner.ScanSPVec3()
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func Normals(doc *gltf.Document, primitive *gltf.Primitive) ([]sprec.Vec3, error) {
	if !HasAttribute(primitive, gltf.NORMAL) {
		return nil, nil
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.NORMAL]]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	if accessor.Type != gltf.AccessorVec3 {
		return nil, fmt.Errorf("unsupported accessor type %d", accessor.Type)
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([]sprec.Vec3, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		for i := range result {
			result[i] = scanner.ScanSPVec3()
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func Tangents(doc *gltf.Document, primitive *gltf.Primitive) ([]sprec.Vec3, error) {
	if !HasAttribute(primitive, gltf.TANGENT) {
		return nil, nil
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.TANGENT]]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	if accessor.Type != gltf.AccessorVec3 {
		return nil, fmt.Errorf("unsupported accessor type %d", accessor.Type)
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([]sprec.Vec3, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		for i := range result {
			result[i] = scanner.ScanSPVec3()
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func TexCoord0s(doc *gltf.Document, primitive *gltf.Primitive) ([]sprec.Vec2, error) {
	if !HasAttribute(primitive, gltf.TEXCOORD_0) {
		return nil, nil
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.TEXCOORD_0]]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	if accessor.Type != gltf.AccessorVec2 {
		return nil, fmt.Errorf("unsupported accessor type %d", accessor.Type)
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([]sprec.Vec2, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		for i := range result {
			result[i] = scanner.ScanSPVec2()
			result[i].Y = 1.0 - result[i].Y // fix tex coord orientation
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func Color0s(doc *gltf.Document, primitive *gltf.Primitive) ([]sprec.Vec4, error) {
	if !HasAttribute(primitive, gltf.COLOR_0) {
		return nil, nil
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.COLOR_0]]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	if accessor.Type != gltf.AccessorVec4 {
		return nil, fmt.Errorf("unsupported accessor type %d", accessor.Type)
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([]sprec.Vec4, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		for i := range result {
			result[i] = scanner.ScanSPVec4()
		}
	case gltf.ComponentUshort:
		for i := range result {
			r := scanner.ScanUint16()
			g := scanner.ScanUint16()
			b := scanner.ScanUint16()
			a := scanner.ScanUint16()
			result[i] = sprec.NewVec4(
				float32(r)/float32(0xFFFF),
				float32(g)/float32(0xFFFF),
				float32(b)/float32(0xFFFF),
				float32(a)/float32(0xFFFF),
			)
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func Weight0s(doc *gltf.Document, primitive *gltf.Primitive) ([]sprec.Vec4, error) {
	if !HasAttribute(primitive, gltf.WEIGHTS_0) {
		return nil, nil
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.WEIGHTS_0]]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	if accessor.Type != gltf.AccessorVec4 {
		return nil, fmt.Errorf("unsupported accessor type %d", accessor.Type)
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([]sprec.Vec4, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		for i := range result {
			result[i] = scanner.ScanSPVec4()
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func Joint0s(doc *gltf.Document, primitive *gltf.Primitive) ([][4]uint8, error) {
	if !HasAttribute(primitive, gltf.JOINTS_0) {
		return nil, nil
	}
	accessor := doc.Accessors[primitive.Attributes[gltf.JOINTS_0]]
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor lacks a buffer view")
	}
	if accessor.Type != gltf.AccessorVec4 {
		return nil, fmt.Errorf("unsupported accessor type %d", accessor.Type)
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	scanner := blob.NewScanner(buffer)

	result := make([][4]uint8, accessor.Count)
	switch accessor.ComponentType {
	case gltf.ComponentUbyte:
		for i := range result {
			result[i] = [4]uint8{
				scanner.ScanUint8(),
				scanner.ScanUint8(),
				scanner.ScanUint8(),
				scanner.ScanUint8(),
			}
		}
	case gltf.ComponentUshort:
		for i := range result {
			result[i] = [4]uint8{
				uint8(scanner.ScanUint16()),
				uint8(scanner.ScanUint16()),
				uint8(scanner.ScanUint16()),
				uint8(scanner.ScanUint16()),
			}
		}
	default:
		return nil, fmt.Errorf("unsupported accessor component type %d", accessor.ComponentType)
	}
	return result, nil
}

func PrimitiveMaterial(doc *gltf.Document, primitive *gltf.Primitive) *gltf.Material {
	if primitive.Material == nil {
		return nil
	}
	return doc.Materials[*primitive.Material]
}

func BaseColor(pbr *gltf.PBRMetallicRoughness) sprec.Vec4 {
	factor := pbr.BaseColorFactorOrDefault()
	return sprec.NewVec4(float32(factor[0]), float32(factor[1]), float32(factor[2]), float32(factor[3]))
}

func ColorTextureIndex(doc *gltf.Document, pbr *gltf.PBRMetallicRoughness) *uint32 {
	colorTexture := pbr.BaseColorTexture
	if colorTexture == nil {
		return nil
	}
	if colorTexture.TexCoord != 0 {
		logger.Warn("Unsupported color texture: tex coord layer unsupported!")
		return nil
	}
	return &colorTexture.Index
}

func MetallicRoughnessTextureIndex(doc *gltf.Document, pbr *gltf.PBRMetallicRoughness) *uint32 {
	mrTexture := pbr.MetallicRoughnessTexture
	if mrTexture == nil {
		return nil
	}
	if mrTexture.TexCoord != 0 {
		logger.Warn("Unsupported metallic-roughness texture: tex coord layer unsupported!")
		return nil
	}
	return &mrTexture.Index
}

func NormalTextureIndexScale(doc *gltf.Document, material *gltf.Material) (*uint32, float32) {
	normalTexture := material.NormalTexture
	if normalTexture == nil {
		return nil, 1.0
	}
	if normalTexture.TexCoord > 0 {
		logger.Warn("Unsupported normal texture: tex coord layer unsupported!")
		return nil, 1.0
	}
	return normalTexture.Index, float32(normalTexture.ScaleOrDefault())
}

func InverseBindMatrix(doc *gltf.Document, skin *gltf.Skin, index int) sprec.Mat4 {
	if skin.InverseBindMatrices == nil {
		logger.Warn("Skin lacks inverse bind matrices!")
		return sprec.IdentityMat4()
	}

	accessor := doc.Accessors[*skin.InverseBindMatrices]
	if accessor.BufferView == nil {
		logger.Warn("Accessor lacks a buffer view!")
		return sprec.IdentityMat4()
	}

	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorMat4 {
		logger.Error("Unsupported joints accessor type (%d)!", accessor.Type)
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
		logger.Error("Unsupported joints accessor component type (%d)!", accessor.ComponentType)
		return sprec.IdentityMat4()
	}
}

func AnimationKeyframes(doc *gltf.Document, sampler *gltf.AnimationSampler) []float64 {
	accessor := doc.Accessors[sampler.Input]
	if accessor.BufferView == nil {
		logger.Warn("Accessor lacks a buffer view!")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorScalar {
		logger.Error("Unsupported sampler input accessor type (%d)!", accessor.Type)
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
		logger.Error("Unsupported sampler input accessor component type (%d)!", accessor.ComponentType)
		return nil
	}
}

func AnimationTranslations(doc *gltf.Document, sampler *gltf.AnimationSampler) []dprec.Vec3 {
	accessor := doc.Accessors[sampler.Output]
	if accessor.BufferView == nil {
		logger.Warn("Accessor lacks a buffer view!")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorVec3 {
		logger.Error("Unsupported sampler output accessor type (%d)!", accessor.Type)
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
		logger.Error("Unsupported sampler output accessor component type (%d)!", accessor.ComponentType)
		return nil
	}
}

func AnimationRotations(doc *gltf.Document, sampler *gltf.AnimationSampler) []dprec.Quat {
	accessor := doc.Accessors[sampler.Output]
	if accessor.BufferView == nil {
		logger.Warn("Accessor lacks a buffer view!")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorVec4 {
		logger.Error("Unsupported sampler output accessor type (%d)!", accessor.Type)
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
		logger.Error("Unsupported sampler output accessor component type (%d)!", accessor.ComponentType)
		return nil
	}
}

func AnimationScales(doc *gltf.Document, sampler *gltf.AnimationSampler) []dprec.Vec3 {
	accessor := doc.Accessors[sampler.Output]
	if accessor.BufferView == nil {
		logger.Warn("Accessor lacks a buffer view!")
		return nil
	}
	buffer := BufferViewData(doc, *accessor.BufferView)
	if accessor.Type != gltf.AccessorVec3 {
		logger.Error("Unsupported sampler output accessor type (%d)!", accessor.Type)
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
		logger.Error("Unsupported sampler output accessor component type (%d)!", accessor.ComponentType)
		return nil
	}
}

func BufferViewData(doc *gltf.Document, index uint32) gblob.LittleEndianBlock {
	bufferView := doc.BufferViews[index]
	offset := bufferView.ByteOffset
	count := bufferView.ByteLength
	buffer := doc.Buffers[bufferView.Buffer]
	return gblob.LittleEndianBlock(buffer.Data[offset : offset+count])
}

func Properties(extras any) map[string]string {
	result := make(map[string]string)
	props, ok := extras.(map[string]any)
	if ok {
		for key, value := range props {
			if strValue, ok := value.(string); ok {
				result[key] = strValue
			} else {
				result[key] = ""
			}
		}
	}
	return result
}

func IsMeshCollisionDisabled(mesh *gltf.Mesh) bool {
	props, ok := mesh.Extras.(map[string]any)
	if !ok {
		return false
	}
	return props["collision"] == "none"
}

func IsCollisionDisabled(node *gltf.Node) bool {
	props, ok := node.Extras.(map[string]any)
	if !ok {
		return false
	}
	return props["collision"] == "none"
}
