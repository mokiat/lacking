package glsl

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

func (t *Translator) translateGeometryVertexCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties GeometryProperties
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
	return construct("geometry.vert.glsl", properties)
}

func (t *Translator) translateGeometryFragmentCode(shader *lsl.Shader, settings graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties GeometryProperties
	properties.VersionProperties = t.buildVersionProperties()
	properties.AttributeProperties = t.buildAttributeProperties(settings)
	properties.OutputProperties = t.buildOutputProperties(settings)
	properties.TextureProperties = t.buildTextureProperties(ctx, shader)
	properties.UniformProperties = t.buildUniformProperties(ctx, shader)
	properties.VaryingProperties = t.buildVaryingProperties(ctx, shader, "in")
	{
		ctx.Push()
		ctx.RegisterIdentifier("#normal", "normal")
		ctx.RegisterIdentifier("#color", "color")
		ctx.RegisterIdentifier("#metallic", "metallic")
		ctx.RegisterIdentifier("#roughness", "roughness")
		ctx.RegisterIdentifier("#vertexColor", "vertex_color")
		ctx.RegisterIdentifier("#vertexUV", "tex_coord")
		properties.MainProperties = t.buildMainProperties(ctx, shader, "#fragment")
		ctx.Pop()
	}
	return construct("geometry.frag.glsl", properties)
}
