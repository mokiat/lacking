package glsl

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

func (t *Translator) translateForwardVertexCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties ForwardProperties
	properties.VersionProperties = t.buildVersionProperties()
	properties.AttributeProperties = t.buildAttributeProperties(settings)
	properties.OutputProperties = t.buildOutputProperties(settings)
	properties.TextureProperties = t.buildTextureProperties(ctx, shader)
	properties.UniformProperties = t.buildUniformProperties(ctx, shader)
	properties.VaryingProperties = t.buildVaryingProperties(ctx, shader, "out")
	{
		ctx.Push()
		properties.MainProperties = t.buildMainProperties(ctx, shader, "#vertex")
		ctx.Pop()
	}
	return construct("forward.vert.glsl", properties)
}

func (t *Translator) translateForwardFragmentCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties ForwardProperties
	properties.VersionProperties = t.buildVersionProperties()
	properties.AttributeProperties = t.buildAttributeProperties(settings)
	properties.OutputProperties = t.buildOutputProperties(settings)
	properties.TextureProperties = t.buildTextureProperties(ctx, shader)
	properties.UniformProperties = t.buildUniformProperties(ctx, shader)
	properties.VaryingProperties = t.buildVaryingProperties(ctx, shader, "in")
	{
		ctx.Push()
		ctx.RegisterIdentifier("#color", "color")
		properties.MainProperties = t.buildMainProperties(ctx, shader, "#fragment")
		ctx.Pop()
	}
	return construct("forward.frag.glsl", properties)
}
