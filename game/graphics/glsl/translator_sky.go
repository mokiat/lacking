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

	var properties SkyProperties
	properties.BaseProperties = t.buildBaseProperties(ctx, shader, constraints, "in")
	{
		ctx.Push()
		// input
		ctx.RegisterIdentifier("#direction", "varyingDirectionWS")
		// camera
		ctx.RegisterIdentifier("#cameraMatrix", "cameraMatrixIn")
		ctx.RegisterIdentifier("#viewMatrix", "viewMatrixIn")
		ctx.RegisterIdentifier("#projectionMatrix", "projectionMatrixIn")
		ctx.RegisterIdentifier("#viewport", "viewportIn")
		// output
		ctx.RegisterIdentifier("#color", "color")

		properties.MainProperties = t.buildMainProperties(ctx, shader, "#fragment")
		ctx.Pop()
	}
	return construct("sky.frag.glsl", properties)
}
