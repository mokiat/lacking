package glsl

import (
	_ "embed"
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

func NewTranslator(version string, precisionQualifiers bool) *Translator {
	return &Translator{
		version:                version,
		hasPrecisionQualifiers: precisionQualifiers,
	}
}

type Translator struct {
	version                string
	hasPrecisionQualifiers bool
}

func (t *Translator) Translate(shader *lsl.Shader, settings graphics.ShaderConstraints) ProgramCode {
	switch settings.Preset {
	case graphics.PresetSky:
		return ProgramCode{
			VertexCode:   t.translateSkyVertexCode(shader, settings),
			FragmentCode: t.translateSkyFragmentCode(shader, settings),
		}
	default:
		panic(fmt.Errorf("unsupported shader preset: %s", settings.Preset))
	}
}

func (t *Translator) translateSkyVertexCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties SkyProperties
	properties.VersionProperties = t.buildVersionProperties()
	properties.AttributeProperties = t.buildAttributeProperties(settings)
	properties.OutputProperties = t.buildOutputProperties(settings)
	properties.TextureProperties = t.buildTextureProperties(ctx, shader)
	properties.UniformProperties = t.buildUniformProperties(ctx, shader)
	properties.VaryingProperties = t.buildVaryingProperties(ctx, shader, "out")
	return construct("sky.vert.glsl", properties)
}

func (t *Translator) translateSkyFragmentCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties SkyProperties
	properties.VersionProperties = t.buildVersionProperties()
	properties.AttributeProperties = t.buildAttributeProperties(settings)
	properties.OutputProperties = t.buildOutputProperties(settings)
	properties.TextureProperties = t.buildTextureProperties(ctx, shader)
	properties.UniformProperties = t.buildUniformProperties(ctx, shader)
	properties.VaryingProperties = t.buildVaryingProperties(ctx, shader, "in")
	return construct("sky.frag.glsl", properties)
}

func (t *Translator) buildVersionProperties() VersionProperties {
	return VersionProperties{
		Version:        t.version,
		NeedsPrecision: t.hasPrecisionQualifiers,
	}
}

func (t *Translator) buildAttributeProperties(settings graphics.ShaderConstraints) AttributeProperties {
	return AttributeProperties{
		HasAttributeCoord:    settings.HasCoords,
		HasAttributeNormal:   settings.HasNormals,
		HasAttributeTangent:  settings.HasTangents,
		HasAttributeTexCoord: settings.HasTexCoords,
		HasAttributeColor:    settings.HasVertexColors,
		HasAttributeArmature: settings.HasArmature,
	}
}

func (t *Translator) buildOutputProperties(settings graphics.ShaderConstraints) OutputProperties {
	return OutputProperties{
		HasFramebufferOutput0: settings.HasOutput0,
		HasFramebufferOutput1: settings.HasOutput1,
		HasFramebufferOutput2: settings.HasOutput2,
		HasFramebufferOutput3: settings.HasOutput3,
	}
}

func (t *Translator) buildTextureProperties(ctx *translationContext, shader *lsl.Shader) TextureProperties {
	return TextureProperties{
		Textures: gog.MapIndex(shader.Textures(), func(index int, texture lsl.Field) TextureProperty {
			srcName := texture.Name
			dstName := fmt.Sprintf("uTexture%d", index)
			dstType := t.translateType(ctx, texture.Type)
			ctx.RegisterMapping(srcName, dstName)
			return TextureProperty{
				Name: dstName,
				Type: dstType,
			}
		}),
	}
}

func (t *Translator) buildUniformProperties(ctx *translationContext, shader *lsl.Shader) UniformProperties {
	return UniformProperties{
		Uniforms: gog.MapIndex(shader.Uniforms(), func(index int, uniform lsl.Field) UniformProperty {
			srcName := uniform.Name
			dstName := fmt.Sprintf("uUniform%d", index)
			dstType := t.translateType(ctx, uniform.Type)
			ctx.RegisterMapping(srcName, dstName)
			return UniformProperty{
				Name: dstName,
				Type: dstType,
			}
		}),
	}
}

func (t *Translator) buildVaryingProperties(ctx *translationContext, shader *lsl.Shader, direction string) VaryingProperties {
	return VaryingProperties{
		Varyings: gog.MapIndex(shader.Varyings(), func(index int, varying lsl.Field) VaryingProperty {
			srcName := varying.Name
			dstName := fmt.Sprintf("uVarying%d", index)
			dstType := t.translateType(ctx, varying.Type)
			ctx.RegisterMapping(srcName, dstName)
			return VaryingProperty{
				Name:      dstName,
				Type:      dstType,
				Direction: direction,
			}
		}),
	}
}

func (t *Translator) translateType(_ *translationContext, srcType string) string {
	switch srcType {
	case lsl.TypeNameBool:
		return "bool"
	case lsl.TypeNameInt:
		return "int"
	case lsl.TypeNameUint:
		return "uint"
	case lsl.TypeNameFloat:
		return "float"
	case lsl.TypeNameVec2:
		return "vec2"
	case lsl.TypeNameVec3:
		return "vec3"
	case lsl.TypeNameVec4:
		return "vec4"
	case lsl.TypeNameBVec2:
		return "bvec2"
	case lsl.TypeNameBVec3:
		return "bvec3"
	case lsl.TypeNameBVec4:
		return "bvec4"
	case lsl.TypeNameIVec2:
		return "ivec2"
	case lsl.TypeNameIVec3:
		return "ivec3"
	case lsl.TypeNameIVec4:
		return "ivec4"
	case lsl.TypeNameUVec2:
		return "uvec2"
	case lsl.TypeNameUVec3:
		return "uvec3"
	case lsl.TypeNameUVec4:
		return "uvec4"
	case lsl.TypeNameMat2:
		return "mat2"
	case lsl.TypeNameMat3:
		return "mat3"
	case lsl.TypeNameMat4:
		return "mat4"
	case lsl.TypeNameSampler2D:
		return "sampler2D"
	case lsl.TypeNameSamplerCube:
		return "samplerCube"
	default:
		panic(fmt.Errorf("unknown type name: %s", srcType))
	}
}
