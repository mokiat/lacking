package glsl

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

func (t *Translator) translateShadowVertexCode(shader *lsl.Shader, constraints graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties ShadowProperties
	properties.BaseProperties = t.buildBaseProperties(ctx, shader, constraints, "out")
	{
		ctx.Push()
		properties.MainProperties = t.buildMainProperties(ctx, shader, "#vertex")
		ctx.Pop()
	}
	return construct("shadow.vert.glsl", properties)
}

func (t *Translator) translateShadowFragmentCode(shader *lsl.Shader, constraints graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties ShadowProperties
	properties.BaseProperties = t.buildBaseProperties(ctx, shader, constraints, "in")
	{
		ctx.Push()
		properties.MainProperties = t.buildMainProperties(ctx, shader, "#fragment")
		ctx.Pop()
	}
	return construct("shadow.frag.glsl", properties)
}
