package internal

import "github.com/mokiat/lacking/framework/opengl"

func NewPatchDrawMaterial() *Material {
	vs := func() string {
		builder := opengl.NewShaderSourceBuilder(patchDrawMaterialVertexShaderTemplate)
		return builder.Build()
	}
	tcs := func() string {
		builder := opengl.NewShaderSourceBuilder(patchDrawMaterialTesselationControlShaderTemplate)
		return builder.Build()
	}
	tes := func() string {
		builder := opengl.NewShaderSourceBuilder(patchDrawMaterialTesselationEvaluationShaderTemplate)
		return builder.Build()
	}
	fs := func() string {
		builder := opengl.NewShaderSourceBuilder(patchDrawMaterialFragmentShaderTemplate)
		return builder.Build()
	}
	return newTessMaterial(vs, tcs, tes, fs)
}

const patchDrawMaterialVertexShaderTemplate = `
layout(location = 0) in vec2 positionIn;

uniform mat4 projectionMatrixIn;

out vec4 positionCS;

void main()
{
	positionCS = projectionMatrixIn * vec4(positionIn, 0.0, 1.0);
}
`

const patchDrawMaterialTesselationControlShaderTemplate = `
layout (vertices = 5) out;

in vec4 positionCS[];

out vec4 positionEval[];

void main()
{
	positionEval[gl_InvocationID] = positionCS[gl_InvocationID];
	gl_TessLevelOuter[0] = 1;
	gl_TessLevelOuter[1] = 30;
	gl_TessLevelOuter[2] = 1;
	gl_TessLevelInner[0] = 1;
}
`

const patchDrawMaterialTesselationEvaluationShaderTemplate = `
layout(triangles, equal_spacing, ccw) in;

in vec4 positionEval[];

void main()
{
	if (gl_TessCoord.y > 0.0) {
		gl_Position = positionEval[0] * gl_TessCoord.y + positionEval[1] * gl_TessCoord.z + positionEval[2] * gl_TessCoord.x;
	} else {
		float t = gl_TessCoord.x;
		gl_Position = positionEval[3] + (1-t)*(1-t)*(positionEval[1]-positionEval[3]) + t*t*(positionEval[2]-positionEval[3]);
	}
}
`

const patchDrawMaterialFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

uniform sampler2D textureIn;

void main()
{
	fragmentColor = vec4(1.0, 0.0, 0.0, 1.0);
}
`
