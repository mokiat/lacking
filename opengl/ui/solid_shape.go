package ui

const solidShapeVertexShaderTemplate = `
layout(location = 0) in vec3 positionIn;
layout(location = 1) in vec2 texCoordIn;
layout(location = 2) in vec4 colorIn;

uniform mat4 projectionMatrixIn;

noperspective out vec4 colorInOut;

void main()
{
	colorInOut = colorIn;
	gl_Position = projectionMatrixIn * vec4(positionIn, 1.0);
}
`

const solidShapeFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

noperspective in vec4 colorInOut;

void main()
{
	fragmentColor = colorInOut;
}
`
