package glsl

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

func (t *Translator) translateSkyVertexCode(shader *lsl.Shader, constraints graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties SkyProperties
	properties.BaseProperties = t.buildBaseProperties(ctx, shader, constraints, "out")
	{
		ctx.Push()
		properties.MainProperties = t.buildMainProperties(ctx, shader, "#vertex")
		ctx.Pop()
	}
	return construct("sky.vert.glsl", properties)
}

func (t *Translator) translateSkyFragmentCode(shader *lsl.Shader, constraints graphics.ShaderConstraints) string {
	ctx := newTranslationContext()
	ctx.RegisterIdentifier("#rayDirectionWS", "varyingDirectionWS")

	var properties SkyProperties
	properties.BaseProperties = t.buildBaseProperties(ctx, shader, constraints, "in")
	{
		ctx.Push()
		ctx.RegisterIdentifier("#color", "color")
		properties.MainProperties = t.buildMainProperties(ctx, shader, "#fragment")
		ctx.Pop()
	}
	return construct("sky.frag.glsl", properties)
}
