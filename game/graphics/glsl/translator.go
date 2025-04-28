package glsl

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

var (
	//go:embed snippet/precision.glsl
	precisionSnippet string
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
	return ProgramCode{
		VertexCode:   t.translateVertexCode(shader, settings),
		FragmentCode: t.translateFragmentCode(shader, settings),
	}
}

func (t *Translator) translateVertexCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var builder strings.Builder
	t.writeVersion(&builder)
	if settings.HasCoords {
		t.writeCoordAttribute(&builder)
	}
	if settings.HasNormals {
		t.writeNormalAttribute(&builder)
	}
	if settings.HasTangents {
		t.writeTangentAttribute(&builder)
	}
	if settings.HasTexCoords {
		t.writeTexCoordAttribute(&builder)
	}
	if settings.HasVertexColors {
		t.writeVertexColorAttribute(&builder)
	}
	if settings.HasArmature {
		t.writeArmatureAttributes(&builder)
	}
	if textures := shader.Textures(); len(textures) > 0 {
		t.writeTextures(&builder, ctx, textures)
	}
	return builder.String()
}

func (t *Translator) translateFragmentCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var builder strings.Builder
	t.writeVersion(&builder)
	if t.hasPrecisionQualifiers {
		t.writePrecision(&builder)
	}
	if textures := shader.Textures(); len(textures) > 0 {
		t.writeTextures(&builder, ctx, textures)
	}
	return builder.String()
}

func (t *Translator) writeVersion(builder *strings.Builder) {
	fmt.Fprintf(builder, "#version %s\n\n", t.version)
}

func (t *Translator) writePrecision(builder *strings.Builder) {
	builder.WriteString(precisionSnippet)
	builder.WriteString("\n\n")
}

func (t *Translator) writeCoordAttribute(builder *strings.Builder) {
	builder.WriteString("layout(location = 0) in vec4 attrCoord;\n")
}

func (t *Translator) writeNormalAttribute(builder *strings.Builder) {
	builder.WriteString("layout(location = 1) in vec3 attrNormal;\n")
}

func (t *Translator) writeTangentAttribute(builder *strings.Builder) {
	builder.WriteString("layout(location = 2) in vec3 attrTangent;\n")
}

func (t *Translator) writeTexCoordAttribute(builder *strings.Builder) {
	builder.WriteString("layout(location = 3) in vec2 attrTexCoord;\n")
}

func (t *Translator) writeVertexColorAttribute(builder *strings.Builder) {
	builder.WriteString("layout(location = 4) in vec4 attrColor;\n")
}

func (t *Translator) writeArmatureAttributes(builder *strings.Builder) {
	builder.WriteString("layout(location = 5) in vec4 attrWeights;\n")
	builder.WriteString("layout(location = 6) in uvec4 attrJoints;\n")
}

func (t *Translator) writeTextures(builder *strings.Builder, ctx *translationContext, textures []lsl.Field) {
	for i, texture := range textures {
		srcName := texture.Name
		dstName := fmt.Sprintf("uTexture%d", i)
		dstType := t.translateType(ctx, texture.Type)
		ctx.RegisterMapping(srcName, dstName)
		fmt.Fprintf(builder, "uniform %s %s;\n", dstType, dstName)
	}
	builder.WriteString("\n")
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
