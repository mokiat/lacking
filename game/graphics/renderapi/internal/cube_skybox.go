package internal

import (
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/render"
)

func NewCubeSkyboxPresentation(api render.API) *SkyboxPresentation {
	vsBuilder := opengl.NewShaderSourceBuilder(cubeSkyboxVertexShader)
	fsBuilder := opengl.NewShaderSourceBuilder(cubeSkyboxFragmentShader)
	return NewSkyboxPresentation(api, vsBuilder.Build(), fsBuilder.Build())
}

const cubeSkyboxVertexShader = `
layout(location = 0) in vec3 coordIn;

uniform mat4 projectionMatrixIn;
uniform mat4 viewMatrixIn;

smooth out vec3 texCoordInOut;

void main()
{
	// we optimize by using vertex coords as cube texture coords
	// additionally, we need to flip the coords. opengl uses renderman coordinate
	// system for cube maps, contrary to the rest of the opengl api
	texCoordInOut = -coordIn;

	// ensure that translations are ignored by setting w to 0.0
	vec4 viewPosition = viewMatrixIn * vec4(coordIn, 0.0);

	// restore w to 1.0 so that projection works
	vec4 position = projectionMatrixIn * vec4(viewPosition.xyz, 1.0);

	// set z to w so that it has maximum depth (1.0) after projection division
	gl_Position = vec4(position.xy, position.w, position.w);
}`

const cubeSkyboxFragmentShader = `
layout(location = 0) out vec4 fbColor0Out;

uniform samplerCube albedoCubeTextureIn;

smooth in vec3 texCoordInOut;

void main()
{
	fbColor0Out = texture(albedoCubeTextureIn, texCoordInOut);
}`

func NewColorSkyboxPresentation(api render.API) *SkyboxPresentation {
	vsBuilder := opengl.NewShaderSourceBuilder(colorSkyboxVertexShader)
	fsBuilder := opengl.NewShaderSourceBuilder(colorSkyboxFragmentShader)
	return NewSkyboxPresentation(api, vsBuilder.Build(), fsBuilder.Build())
}

const colorSkyboxVertexShader = `
layout(location = 0) in vec3 coordIn;

uniform mat4 projectionMatrixIn;
uniform mat4 viewMatrixIn;

void main()
{
	// ensure that translations are ignored by setting w to 0.0
	vec4 viewPosition = viewMatrixIn * vec4(coordIn, 0.0);

	// restore w to 1.0 so that projection works
	vec4 position = projectionMatrixIn * vec4(viewPosition.xyz, 1.0);

	// set z to w so that it has maximum depth (1.0) after projection division
	gl_Position = vec4(position.xy, position.w, position.w);
}`

const colorSkyboxFragmentShader = `
layout(location = 0) out vec4 fbColor0Out;

uniform vec4 albedoColorIn;

void main()
{
	fbColor0Out = albedoColorIn;
}`
