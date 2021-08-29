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
uniform vec4 clipDistancesIn;

out gl_PerVertex
{
  vec4 gl_Position;
  float gl_ClipDistance[4];
};

noperspective out vec4 colorInOut;
noperspective out vec2 texCoordInOut;

void main()
{
	texCoordInOut = texCoordIn;
	colorInOut = colorIn;
	gl_ClipDistance[0] = positionIn.x - clipDistancesIn.x; // left
	gl_ClipDistance[1] = clipDistancesIn.y - positionIn.x; // right
	gl_ClipDistance[2] = positionIn.y - clipDistancesIn.z; // top
	gl_ClipDistance[3] = clipDistancesIn.w - positionIn.y; // bottom
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
