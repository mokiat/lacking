package internal

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

const (
	PropertyKindUnknown PropertyKind = iota
	PropertyKindFloat
	PropertyKindVec2
	PropertyKindVec3
	PropertyKindVec4
)

type PropertyKind uint8

func ResolvePropertyKind(value any) PropertyKind {
	switch value.(type) {
	case float32:
		return PropertyKindFloat
	case sprec.Vec2:
		return PropertyKindVec2
	case sprec.Vec3:
		return PropertyKindVec3
	case sprec.Vec4:
		return PropertyKindVec4
	default:
		return PropertyKindUnknown
	}
}

type UniformProperty struct {
	name       string
	kind       PropertyKind
	byteOffset uint32
}

func NewShaderUniformSet(shader *lsl.Shader) UniformSet {
	uniformBlock, ok := shader.FindUniformBlock()
	if !ok {
		return UniformSet{}
	}

	padToMultiple := func(value, multiple uint32) uint32 {
		if value%multiple > 0 {
			return value + (multiple - value%multiple)
		}
		return value
	}

	size := uint32(0)
	properties := make([]UniformProperty, len(uniformBlock.Fields))
	for i, field := range uniformBlock.Fields {
		property := UniformProperty{
			name: field.Name,
		}
		switch field.Type {
		case lsl.TypeNameFloat:
			property.kind = PropertyKindFloat
			size = padToMultiple(size, 4)
			property.byteOffset = size
			size += 4
		case lsl.TypeNameVec2:
			property.kind = PropertyKindVec2
			size = padToMultiple(size, 8)
			property.byteOffset = size
			size += 8
		case lsl.TypeNameVec3:
			property.kind = PropertyKindVec3
			size = padToMultiple(size, 16) // vec3 has weird padding
			property.byteOffset = size
			size += 12
		case lsl.TypeNameVec4:
			property.kind = PropertyKindVec4
			size = padToMultiple(size, 16)
			property.byteOffset = size
			size += 16
		default:
			panic(fmt.Errorf("unexpected uniform field type: %s", field.Type))
		}
		properties[i] = property
	}

	return UniformSet{
		properties: properties,
		data:       make([]byte, size),
	}
}

type UniformSet struct {
	properties []UniformProperty
	data       []byte
}

func (s *UniformSet) Property(name string) any {
	if prop, ok := s.findProperty(name); ok {
		block := gblob.LittleEndianBlock(s.data[prop.byteOffset:])
		switch prop.kind {
		case PropertyKindFloat:
			return block.Float32(0)
		case PropertyKindVec2:
			return sprec.Vec2{
				X: block.Float32(0),
				Y: block.Float32(4),
			}
		case PropertyKindVec3:
			return sprec.Vec3{
				X: block.Float32(0),
				Y: block.Float32(4),
				Z: block.Float32(8),
			}
		case PropertyKindVec4:
			return sprec.Vec4{
				X: block.Float32(0),
				Y: block.Float32(4),
				Z: block.Float32(8),
				W: block.Float32(12),
			}
		default:
			panic(fmt.Errorf("unexpected property kind: %d", prop.kind))
		}
	}
	return nil
}

func (s *UniformSet) SetProperty(name string, value any) {
	if prop, ok := s.findProperty(name); ok {
		kind := ResolvePropertyKind(value)
		block := gblob.LittleEndianBlock(s.data[prop.byteOffset:])
		switch value := value.(type) {
		case []byte:
			copy(block, value)
		case float32:
			if kind == PropertyKindFloat {
				block.SetFloat32(0, value)
			}
		case sprec.Vec2:
			if kind == PropertyKindVec2 {
				block.SetFloat32(0, value.X)
				block.SetFloat32(4, value.Y)
			}
		case sprec.Vec3:
			if kind == PropertyKindVec3 {
				block.SetFloat32(0, value.X)
				block.SetFloat32(4, value.Y)
				block.SetFloat32(8, value.Z)
			}
		case sprec.Vec4:
			if kind == PropertyKindVec4 {
				block.SetFloat32(0, value.X)
				block.SetFloat32(4, value.Y)
				block.SetFloat32(8, value.Z)
				block.SetFloat32(12, value.W)
			}
		default:
			panic(fmt.Errorf("unexpected property type: %T", value))
		}
	}
}

func (s *UniformSet) Data() []byte {
	return s.data
}

func (s *UniformSet) IsEmpty() bool {
	return len(s.properties) == 0 || len(s.data) == 0
}

func (s *UniformSet) findProperty(name string) (UniformProperty, bool) {
	return gog.FindFunc(s.properties, func(prop UniformProperty) bool {
		return prop.name == name
	})
}
