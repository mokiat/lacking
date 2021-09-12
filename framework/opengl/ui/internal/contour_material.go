package internal

import "github.com/mokiat/lacking/framework/opengl"

func newContourMaterial() *Material {
	vs := func() string {
		builder := opengl.NewShaderSourceBuilder(contourMaterialVertexShaderTemplate)
		return builder.Build()
	}
	fs := func() string {
		builder := opengl.NewShaderSourceBuilder(contourMaterialFragmentShaderTemplate)
		return builder.Build()
	}
	return newMaterial(vs, fs)
}

const contourMaterialVertexShaderTemplate = `
layout(location = 0) in vec2 positionIn;
layout(location = 2) in vec4 colorIn;

uniform mat4 transformMatrixIn;
uniform mat4 projectionMatrixIn;
uniform vec4 clipDistancesIn;

out gl_PerVertex
{
  vec4 gl_Position;
  float gl_ClipDistance[4];
};

noperspective out vec4 colorInOut;

void main()
{
	colorInOut = colorIn;
	vec4 screenPosition = transformMatrixIn * vec4(positionIn, 0.0, 1.0);
	gl_ClipDistance[0] = screenPosition.x - clipDistancesIn.x; // left
	gl_ClipDistance[1] = clipDistancesIn.y - screenPosition.x; // right
	gl_ClipDistance[2] = screenPosition.y - clipDistancesIn.z; // top
	gl_ClipDistance[3] = clipDistancesIn.w - screenPosition.y; // bottom
	gl_Position = projectionMatrixIn * screenPosition;
}
`

const contourMaterialFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

noperspective in vec4 colorInOut;

void main()
{
	fragmentColor = colorInOut;
}
`
