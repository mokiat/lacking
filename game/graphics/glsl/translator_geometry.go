package glsl

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

func (t *Translator) translateGeometryVertexCode(shader *lsl.Shader, constraints graphics.ShaderConstraints) string {
	ctx := newTranslationContext()
	ctx.RegisterIdentifier("#time", "timeIn")

	var properties GeometryProperties
	properties.BaseProperties = t.buildBaseProperties(ctx, shader, constraints, "out")
	{
		ctx.Push()
		// time
		ctx.RegisterIdentifier("#time", "timeIn")
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
		ctx.RegisterIdentifier("#varyingNormal", "normalInOut")
		ctx.RegisterIdentifier("#varyingTangent", "tangentInOut")
		ctx.RegisterIdentifier("#varyingUV", "texCoordInOut")
		ctx.RegisterIdentifier("#varyingColor", "colorInOut")
		ctx.RegisterIdentifier("#position", "position")

		properties.MainProperties = t.buildMainProperties(ctx, shader, "#vertex")
		ctx.Pop()
	}
	return construct("geometry.vert.glsl", properties)
}

func (t *Translator) translateGeometryFragmentCode(shader *lsl.Shader, constraints graphics.ShaderConstraints) string {
	ctx := newTranslationContext()

	var properties GeometryProperties
	properties.BaseProperties = t.buildBaseProperties(ctx, shader, constraints, "in")
	{
		ctx.Push()
		// time
		ctx.RegisterIdentifier("#time", "timeIn")
		// camera
		ctx.RegisterIdentifier("#cameraMatrix", "cameraMatrixIn")
		ctx.RegisterIdentifier("#viewMatrix", "viewMatrixIn")
		ctx.RegisterIdentifier("#projectionMatrix", "projectionMatrixIn")
		ctx.RegisterIdentifier("#viewport", "viewportIn")
		// input
		ctx.RegisterIdentifier("#varyingNormal", "normalInOut")
		ctx.RegisterIdentifier("#varyingTangent", "tangentInOut")
		ctx.RegisterIdentifier("#varyingUV", "texCoordInOut")
		ctx.RegisterIdentifier("#varyingColor", "colorInOut")
		// output
		ctx.RegisterIdentifier("#normal", "normal_ws")
		ctx.RegisterIdentifier("#color", "color")
		ctx.RegisterIdentifier("#metallic", "metallic")
		ctx.RegisterIdentifier("#roughness", "roughness")

		properties.MainProperties = t.buildMainProperties(ctx, shader, "#fragment")
		ctx.Pop()
	}
	return construct("geometry.frag.glsl", properties)
}
