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
		// time
		ctx.RegisterIdentifier("#time", "timeIn")
		ctx.RegisterIdentifier("#spawnTime", "spawnTimeInOut")
		ctx.RegisterIdentifier("#custom0", "custom0InOut")
		ctx.RegisterIdentifier("#custom1", "custom1InOut")
		ctx.RegisterIdentifier("#custom2", "custom2InOut")
		// mesh
		ctx.RegisterIdentifier("#vertexCoord", "coord_ls")
		ctx.RegisterIdentifier("#vertexNormal", "normal_ls")
		ctx.RegisterIdentifier("#vertexTangent", "tangent_ls")
		ctx.RegisterIdentifier("#vertexUV", "tex_coord")
		ctx.RegisterIdentifier("#vertexColor", "color")
		// model
		ctx.RegisterIdentifier("#modelMatrix", "model_matrix")
		// camera
		ctx.RegisterIdentifier("#cameraMatrix", "cameraMatrixIn")
		ctx.RegisterIdentifier("#viewMatrix", "viewMatrixIn")
		ctx.RegisterIdentifier("#projectionMatrix", "projectionMatrixIn")
		ctx.RegisterIdentifier("#viewport", "viewportIn")
		// output
		ctx.RegisterIdentifier("#position", "position")

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
		// time
		ctx.RegisterIdentifier("#time", "timeIn")
		ctx.RegisterIdentifier("#spawnTime", "spawnTimeInOut")
		ctx.RegisterIdentifier("#custom0", "custom0InOut")
		ctx.RegisterIdentifier("#custom1", "custom1InOut")
		ctx.RegisterIdentifier("#custom2", "custom2InOut")
		// camera
		ctx.RegisterIdentifier("#cameraMatrix", "cameraMatrixIn")
		ctx.RegisterIdentifier("#viewMatrix", "viewMatrixIn")
		ctx.RegisterIdentifier("#projectionMatrix", "projectionMatrixIn")
		ctx.RegisterIdentifier("#viewport", "viewportIn")

		properties.MainProperties = t.buildMainProperties(ctx, shader, "#fragment")
		ctx.Pop()
	}
	return construct("shadow.frag.glsl", properties)
}
