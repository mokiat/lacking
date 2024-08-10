package lsl

import "github.com/mokiat/gog/ds"

type Schema interface {
	// IsAllowedTextureType returns whether the provided type name is
	// allowed to be used in a texture block.
	IsAllowedTextureType(typeName string) bool

	// IsAllowedUniformType returns whether the provided type name is
	// allowed to be used in a uniform block.
	IsAllowedUniformType(typeName string) bool

	// IsAllowedVaryingType returns whether the provided type name is
	// allowed to be used in a varying block.
	IsAllowedVaryingType(typeName string) bool
}

// DefaultSchema returns the default schema implementation.
func DefaultSchema() Schema {
	return defaultSchemaInst
}

var defaultSchemaInst = func() Schema {
	result := &defaultSchema{
		allowedTextureTypes: ds.SetFromSlice([]string{
			TypeNameSampler2D, TypeNameSamplerCube,
		}),
		allowedUniformTypes: ds.SetFromSlice([]string{
			TypeNameBool, TypeNameInt, TypeNameUint, TypeNameFloat,
			TypeNameVec2, TypeNameVec3, TypeNameVec4,
			TypeNameBVec2, TypeNameBVec3, TypeNameBVec4,
			TypeNameIVec2, TypeNameIVec3, TypeNameIVec4,
			TypeNameUVec2, TypeNameUVec3, TypeNameUVec4,
			TypeNameMat2, TypeNameMat3, TypeNameMat4,
		}),
		allowedVaryingTypes: ds.SetFromSlice([]string{
			TypeNameFloat, TypeNameVec2, TypeNameVec3, TypeNameVec4,
		}),
	}
	return result
}()

type defaultSchema struct {
	allowedTextureTypes *ds.Set[string]
	allowedUniformTypes *ds.Set[string]
	allowedVaryingTypes *ds.Set[string]
}

func (s *defaultSchema) IsAllowedTextureType(typeName string) bool {
	return s.allowedTextureTypes.Contains(typeName)
}

func (s *defaultSchema) IsAllowedUniformType(typeName string) bool {
	return s.allowedUniformTypes.Contains(typeName)
}

func (s *defaultSchema) IsAllowedVaryingType(typeName string) bool {
	return s.allowedVaryingTypes.Contains(typeName)
}

func GeometrySchema() Schema {
	return DefaultSchema() // TODO
}

func ShadowSchema() Schema {
	return DefaultSchema() // TODO
}

func ForwardSchema() Schema {
	return DefaultSchema() // TODO
}

func SkySchema() Schema {
	return DefaultSchema() // TODO
}

func PostprocessSchema() Schema {
	return DefaultSchema() // TODO
}
