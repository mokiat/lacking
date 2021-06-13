package internal

import "github.com/mokiat/lacking/framework/opengl"

func NewDrawMaterial() *Material {
	vs := func() string {
		builder := opengl.NewShaderSourceBuilder(drawMaterialVertexShaderTemplate)
		return builder.Build()
	}
	fs := func() string {
		builder := opengl.NewShaderSourceBuilder(drawMaterialFragmentShaderTemplate)
		return builder.Build()
	}
	return newMaterial(vs, fs)
}

const drawMaterialVertexShaderTemplate = `
layout(location = 0) in vec2 positionIn;
layout(location = 1) in vec2 texCoordIn;
layout(location = 2) in vec4 colorIn;

uniform mat4 projectionMatrixIn;

noperspective out vec4 colorInOut;
noperspective out vec2 texCoordInOut;

void main()
{
	texCoordInOut = texCoordIn;
	colorInOut = colorIn;
	gl_Position = projectionMatrixIn * vec4(positionIn, 0.0, 1.0);
}
`

const drawMaterialFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

uniform sampler2D textureIn;

noperspective in vec2 texCoordInOut;
noperspective in vec4 colorInOut;

void main()
{
	fragmentColor = texture(textureIn, texCoordInOut) * colorInOut;
}
`
