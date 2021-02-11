package ui

const imageShapeVertexShaderTemplate = `
layout(location = 0) in vec3 positionIn;
layout(location = 1) in vec2 texCoordIn;

uniform mat4 projectionMatrixIn;

noperspective out vec2 texCoordInOut;

void main()
{
	texCoordInOut = texCoordIn;
	gl_Position = projectionMatrixIn * vec4(positionIn, 1.0);
}
`

const imageShapeFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

uniform sampler2D textureIn;

noperspective in vec2 texCoordInOut;

void main()
{
	vec4 color = texture(textureIn, texCoordInOut);
	if (color.w < 0.5) {
		discard;
	}
	fragmentColor = color;
}
`
