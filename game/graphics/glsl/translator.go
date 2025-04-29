package glsl

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

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
	switch settings.Type {
	case graphics.ShaderTypeShadow:
		return ProgramCode{
			VertexCode:   t.translateShadowVertexCode(shader, settings),
			FragmentCode: t.translateShadowFragmentCode(shader, settings),
		}
	case graphics.ShaderTypeForward:
		return ProgramCode{
			VertexCode:   t.translateForwardVertexCode(shader, settings),
			FragmentCode: t.translateForwardFragmentCode(shader, settings),
		}
	case graphics.ShaderTypeSky:
		return ProgramCode{
			VertexCode:   t.translateSkyVertexCode(shader, settings),
			FragmentCode: t.translateSkyFragmentCode(shader, settings),
		}
	default:
		panic(fmt.Errorf("unsupported shader preset: %s", settings.Type))
	}
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
			ctx.RegisterIdentifier(srcName, dstName)
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
			ctx.RegisterIdentifier(srcName, dstName)
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
			ctx.RegisterIdentifier(srcName, dstName)
			return VaryingProperty{
				Name:      dstName,
				Type:      dstType,
				Direction: direction,
			}
		}),
	}
}

func (t *Translator) buildMainProperties(ctx *translationContext, shader *lsl.Shader, functionName string) MainProperties {
	fn, ok := shader.FindFunction(functionName)
	if !ok {
		return MainProperties{}
	}
	return MainProperties{
		MainStatements: gog.Map(fn.Body, func(statement lsl.Statement) string {
			return t.buildStatement(ctx, statement)
		}),
	}
}

func (t *Translator) buildStatement(ctx *translationContext, statement lsl.Statement) string {
	switch stmt := statement.(type) {
	case *lsl.Discard:
		return t.translateDiscard(ctx, stmt)
	case *lsl.Assignment:
		return t.translateAssignment(ctx, stmt)
	default:
		panic(fmt.Errorf("unknown statement type: %T", statement))
	}
}

func (t *Translator) translateDiscard(_ *translationContext, _ *lsl.Discard) string {
	return "discard;"
}

func (t *Translator) translateAssignment(ctx *translationContext, assignment *lsl.Assignment) string {
	receiver := t.translateExpression(ctx, assignment.Target)
	expression := t.translateExpression(ctx, assignment.Expression)
	operator := t.translateAssignmentOperator(assignment.Operator)
	return fmt.Sprintf("%s %s %s;", receiver, operator, expression)
}

func (t *Translator) translateExpression(ctx *translationContext, expression lsl.Expression) string {
	switch expr := expression.(type) {
	case *lsl.BoolLiteral:
		return t.translateBoolLiteral(expr)
	case *lsl.IntLiteral:
		return t.translateIntLiteral(expr)
	case *lsl.FloatLiteral:
		return t.translateFloatLiteral(expr)
	case *lsl.Identifier:
		return t.translateIdentifier(ctx, expr)
	case *lsl.FieldIdentifier:
		return t.translateFieldIdentifier(ctx, expr)
	case *lsl.UnaryExpression:
		return t.translateUnaryExpression(ctx, expr)
	case *lsl.BinaryExpression:
		return t.translateBinaryExpression(ctx, expr)
	case *lsl.FunctionCall:
		return t.translateFunctionCall(ctx, expr)
	default:
		panic(fmt.Errorf("unknown expression type: %T", expression))
	}
}

func (t *Translator) translateBoolLiteral(literal *lsl.BoolLiteral) string {
	return strconv.FormatBool(literal.Value)
}

func (t *Translator) translateIntLiteral(literal *lsl.IntLiteral) string {
	return strconv.FormatInt(literal.Value, 10)
}

func (t *Translator) translateFloatLiteral(literal *lsl.FloatLiteral) string {
	result := strconv.FormatFloat(literal.Value, 'f', -1, 64)
	if !strings.Contains(result, ".") {
		result += ".0"
	}
	return result
}

func (t *Translator) translateIdentifier(ctx *translationContext, identifier *lsl.Identifier) string {
	return ctx.Identifier(identifier.Name)
}

func (t *Translator) translateFieldIdentifier(ctx *translationContext, identifier *lsl.FieldIdentifier) string {
	owner := t.translateExpression(ctx, identifier.Owner)
	field := identifier.Field.Name // TODO: This needs to be translated too but we need to know the expression type beforehand.
	return fmt.Sprintf("%s.%s", owner, field)
}

func (t *Translator) translateUnaryExpression(ctx *translationContext, expression *lsl.UnaryExpression) string {
	operator := t.translateUnaryOperator(expression.Operator)
	operand := t.translateExpression(ctx, expression.Operand)
	return fmt.Sprintf("%s%s", operator, operand)
}

func (t *Translator) translateBinaryExpression(ctx *translationContext, expression *lsl.BinaryExpression) string {
	left := t.translateExpression(ctx, expression.Left)
	right := t.translateExpression(ctx, expression.Right)
	operator := t.translateBinaryOperator(expression.Operator)
	return fmt.Sprintf("(%s %s %s)", left, operator, right)
}

func (t *Translator) translateFunctionCall(ctx *translationContext, call *lsl.FunctionCall) string {
	identifier := call.Owner.(*lsl.Identifier)
	if _, ok := builtInFunctions[identifier.Name]; ok {
		return t.translateFunctionCallAsIs(ctx, identifier.Name, call.Arguments)
	}
	switch identifier.Name {
	case "sample":
		return t.translateFunctionCallAsIs(ctx, "texture", call.Arguments)
	case "mapNormal":
		return t.translateFunctionCallAsIs(ctx, identifier.Name, call.Arguments)
	default:
		panic(fmt.Errorf("unknown function %q", identifier.Name))
	}
}

func (t *Translator) translateFunctionCallAsIs(ctx *translationContext, name string, arguments []lsl.Expression) string {
	var builder strings.Builder
	builder.WriteString(name)
	builder.WriteString("(")
	lastIndex := len(arguments) - 1
	for i, argument := range arguments {
		builder.WriteString(t.translateExpression(ctx, argument))
		if i != lastIndex {
			builder.WriteString(",")
		}
	}
	builder.WriteString(")")
	return builder.String()
}

func (t *Translator) translateAssignmentOperator(operator string) string {
	switch operator {
	case lsl.AssignmentOperatorEq:
		return "="
	case lsl.AssignmentOperatorAuto:
		panic("auto assignment operator is not supported")
	case lsl.AssignmentOperatorAdd:
		return "+="
	case lsl.AssignmentOperatorSub:
		return "-="
	case lsl.AssignmentOperatorMul:
		return "*="
	case lsl.AssignmentOperatorDiv:
		return "/="
	case lsl.AssignmentOperatorMod:
		return "%="
	case lsl.AssignmentOperatorShl:
		return "<<="
	case lsl.AssignmentOperatorShr:
		return ">>="
	case lsl.AssignmentOperatorAnd:
		return "&="
	case lsl.AssignmentOperatorOr:
		return "|="
	case lsl.AssignmentOperatorXor:
		return "^="
	default:
		panic(fmt.Errorf("unknown assignment operator: %s", operator))
	}
}

func (t *Translator) translateUnaryOperator(operator string) string {
	switch operator {
	case lsl.UnaryOperatorPos:
		return "+"
	case lsl.UnaryOperatorNeg:
		return "-"
	case lsl.UnaryOperatorNot:
		return "!"
	case lsl.UnaryOperatorBitNot:
		return "~" // differs!
	default:
		panic(fmt.Errorf("unknown unary operator: %s", operator))
	}
}

func (t *Translator) translateBinaryOperator(operator string) string {
	switch operator {
	case lsl.BinaryOperatorAdd:
		return "+"
	case lsl.BinaryOperatorSub:
		return "-"
	case lsl.BinaryOperatorMul:
		return "*"
	case lsl.BinaryOperatorDiv:
		return "/"
	case lsl.BinaryOperatorMod:
		return "%"
	case lsl.BinaryOperatorShl:
		return "<<"
	case lsl.BinaryOperatorShr:
		return ">>"
	case lsl.BinaryOperatorEq:
		return "=="
	case lsl.BinaryOperatorNotEq:
		return "!="
	case lsl.BinaryOperatorLess:
		return "<"
	case lsl.BinaryOperatorGreater:
		return ">"
	case lsl.BinaryOperatorLessEq:
		return "<="
	case lsl.BinaryOperatorGreaterEq:
		return ">="
	case lsl.BinaryOperatorBitAnd:
		return "&"
	case lsl.BinaryOperatorBitOr:
		return "|"
	case lsl.BinaryOperatorBitXor:
		return "^"
	case lsl.BinaryOperatorAnd:
		return "&&"
	case lsl.BinaryOperatorOr:
		return "||"
	default:
		panic(fmt.Errorf("unknown binary operator: %s", operator))
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

var builtInFunctions = map[string]struct{}{
	// general functions
	"abs":         {},
	"sign":        {},
	"floor":       {},
	"trunc":       {},
	"round":       {},
	"ceil":        {},
	"fract":       {},
	"min":         {},
	"max":         {},
	"clamp":       {},
	"mix":         {},
	"smoothstep":  {},
	"length":      {},
	"distance":    {},
	"dot":         {},
	"cross":       {},
	"normalize":   {},
	"faceforward": {},
	"reflect":     {},
	"refract":     {},
	"transpose":   {},
	"determinant": {},
	"any":         {},
	"all":         {},
}
