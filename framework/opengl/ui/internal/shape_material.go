package internal

import "github.com/mokiat/lacking/framework/opengl"

func newShapeMaterial() *Material {
	vs := func() string {
		builder := opengl.NewShaderSourceBuilder(shapeMaterialVertexShaderTemplate)
		return builder.Build()
	}
	fs := func() string {
		builder := opengl.NewShaderSourceBuilder(shapeMaterialFragmentShaderTemplate)
		return builder.Build()
	}
	return newMaterial(vs, fs)
}

func newShapeBlankMaterial() *Material {
	vs := func() string {
		builder := opengl.NewShaderSourceBuilder(shapeBlankMaterialVertexShaderTemplate)
		return builder.Build()
	}
	fs := func() string {
		builder := opengl.NewShaderSourceBuilder(shapeBlankMaterialFragmentShaderTemplate)
		return builder.Build()
	}
	return newMaterial(vs, fs)
}

const shapeMaterialVertexShaderTemplate = `
layout(location = 0) in vec2 positionIn;

uniform mat4 transformMatrixIn;
uniform mat4 textureTransformMatrixIn;
uniform mat4 projectionMatrixIn;
uniform vec4 clipDistancesIn;

noperspective out vec2 texCoordInOut;

out gl_PerVertex
{
  vec4 gl_Position;
  float gl_ClipDistance[4];
};

void main()
{
	vec4 screenPosition = transformMatrixIn * vec4(positionIn, 0.0, 1.0);
	texCoordInOut = (textureTransformMatrixIn * vec4(positionIn, 0.0, 1.0)).xy;
	gl_ClipDistance[0] = screenPosition.x - clipDistancesIn.x; // left
	gl_ClipDistance[1] = clipDistancesIn.y - screenPosition.x; // right
	gl_ClipDistance[2] = screenPosition.y - clipDistancesIn.z; // top
	gl_ClipDistance[3] = clipDistancesIn.w - screenPosition.y; // bottom
	gl_Position = projectionMatrixIn * screenPosition;
}
`

const shapeMaterialFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

uniform sampler2D textureIn;
uniform vec4 colorIn = vec4(1.0, 1.0, 1.0, 1.0);

noperspective in vec2 texCoordInOut;

void main()
{
	fragmentColor = texture(textureIn, texCoordInOut) * colorIn;
}
`

const shapeBlankMaterialVertexShaderTemplate = `
layout(location = 0) in vec2 positionIn;

uniform mat4 transformMatrixIn;
uniform mat4 projectionMatrixIn;
uniform vec4 clipDistancesIn;

out gl_PerVertex
{
  vec4 gl_Position;
  float gl_ClipDistance[4];
};

void main()
{
	vec4 screenPosition = transformMatrixIn * vec4(positionIn, 0.0, 1.0);
	gl_ClipDistance[0] = screenPosition.x - clipDistancesIn.x; // left
	gl_ClipDistance[1] = clipDistancesIn.y - screenPosition.x; // right
	gl_ClipDistance[2] = screenPosition.y - clipDistancesIn.z; // top
	gl_ClipDistance[3] = clipDistancesIn.w - screenPosition.y; // bottom
	gl_Position = projectionMatrixIn * screenPosition;
}
`

const shapeBlankMaterialFragmentShaderTemplate = `
layout(location = 0) out vec4 fragmentColor;

void main()
{
	fragmentColor = vec4(1.0, 1.0, 1.0, 1.0);
}
`
